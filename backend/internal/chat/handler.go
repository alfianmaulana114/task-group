package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/alfianhamzah/task-group/backend/internal/auth"
	"github.com/alfianhamzah/task-group/backend/internal/org"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

type sendMessageRequest struct {
	Message          string  `json:"message"`
	MessageType      string  `json:"message_type,omitempty"`
	ReplyToMessageID *string `json:"reply_to_message_id,omitempty"`
}

type chatMessageResponse struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channel_id"`
	OrganizationID   string    `json:"organization_id"`
	ProjectID        string    `json:"project_id"`
	UserID           string    `json:"user_id"`
	Message          string    `json:"message"`
	MessageType      string    `json:"message_type"`
	ReplyToMessageID *string   `json:"reply_to_message_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	User             *userBrief `json:"user,omitempty"`
}

type userBrief struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type channelResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	ProjectID      string `json:"project_id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(auth.ClaimsKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// ensureChannel returns channel id, creating it if not exists (one per project) - race-safe
func (h *Handler) ensureChannel(ctx context.Context, orgID, projectID string) (string, error) {
	var channelID string
	err := h.DB.QueryRow(ctx,
		`SELECT id FROM chat_channels WHERE project_id = $1`, projectID).Scan(&channelID)
	if err == nil {
		return channelID, nil
	}
	if err != pgx.ErrNoRows {
		return "", err
	}
	// create channel with ON CONFLICT to handle TOCTOU
	var name string
	_ = h.DB.QueryRow(ctx, `SELECT name FROM projects WHERE id = $1`, projectID).Scan(&name)
	if name == "" {
		name = "General"
	}
	err = h.DB.QueryRow(ctx,
		`INSERT INTO chat_channels (organization_id, project_id, name, type) VALUES ($1,$2,$3,'project') ON CONFLICT (project_id) DO NOTHING RETURNING id`,
		orgID, projectID, name).Scan(&channelID)
	if err == nil {
		return channelID, nil
	}
	if err == pgx.ErrNoRows {
		// another request created it between SELECT and INSERT
		err = h.DB.QueryRow(ctx, `SELECT id FROM chat_channels WHERE project_id = $1`, projectID).Scan(&channelID)
		return channelID, err
	}
	return "", err
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	orgID := org.GetOrgID(r.Context())
	projectID := chi.URLParam(r, "project_id")
	if orgID == "" || projectID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and project ID required"})
		return
	}
	ctx := r.Context()
	channelID, err := h.ensureChannel(ctx, orgID, projectID)
	if err != nil {
		log.Printf("chat ensure channel: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	rows, err := h.DB.Query(ctx,
		`SELECT cm.id, cm.channel_id, cm.organization_id, cm.project_id, cm.user_id, cm.message, cm.message_type, cm.reply_to_message_id, cm.created_at,
		        u.email, u.full_name
		 FROM chat_messages cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.channel_id = $1 AND cm.deleted_at IS NULL
		 ORDER BY cm.created_at ASC LIMIT 100`, channelID)
	if err != nil {
		log.Printf("chat list: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()
	msgs := []chatMessageResponse{}
	for rows.Next() {
		var m chatMessageResponse
		var u userBrief
		var replyID *string
		var msgType string
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.OrganizationID, &m.ProjectID, &m.UserID, &m.Message, &msgType, &replyID, &m.CreatedAt, &u.Email, &u.FullName); err != nil {
			log.Printf("chat scan: %v", err)
			continue
		}
		m.MessageType = msgType
		m.ReplyToMessageID = replyID
		u.ID = m.UserID
		m.User = &u
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []chatMessageResponse{}
	}
	// also return channel info
	var ch channelResponse
	_ = h.DB.QueryRow(ctx, `SELECT id, organization_id, project_id, name, type FROM chat_channels WHERE id = $1`, channelID).Scan(&ch.ID, &ch.OrganizationID, &ch.ProjectID, &ch.Name, &ch.Type)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"channel":  ch,
		"messages": msgs,
	})
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	orgID := org.GetOrgID(r.Context())
	projectID := chi.URLParam(r, "project_id")
	if orgID == "" || projectID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and project ID required"})
		return
	}
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "message is required"})
		return
	}
	if len(req.Message) > 10000 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "message too long (max 10000 characters)"})
		return
	}
	if req.MessageType == "" {
		req.MessageType = "text"
	}
	if req.MessageType != "text" && req.MessageType != "system" && req.MessageType != "file" {
		req.MessageType = "text"
	}
	ctx := r.Context()
	channelID, err := h.ensureChannel(ctx, orgID, projectID)
	if err != nil {
		log.Printf("chat ensure channel: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	var msg chatMessageResponse
	var replyID interface{} = nil
	if req.ReplyToMessageID != nil && strings.TrimSpace(*req.ReplyToMessageID) != "" {
		replyID = *req.ReplyToMessageID
	}
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		log.Printf("chat begin tx: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()
	err = tx.QueryRow(ctx,
		`INSERT INTO chat_messages (channel_id, organization_id, project_id, user_id, message, message_type, reply_to_message_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 RETURNING id, channel_id, organization_id, project_id, user_id, message, message_type, reply_to_message_id, created_at`,
		channelID, orgID, projectID, claims.UserID, req.Message, req.MessageType, replyID,
	).Scan(&msg.ID, &msg.ChannelID, &msg.OrganizationID, &msg.ProjectID, &msg.UserID, &msg.Message, &msg.MessageType, &msg.ReplyToMessageID, &msg.CreatedAt)
	if err != nil {
		log.Printf("chat insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	// fetch user info (within tx or outside is fine)
	var u userBrief
	_ = tx.QueryRow(ctx, `SELECT email, full_name FROM users WHERE id = $1`, claims.UserID).Scan(&u.Email, &u.FullName)
	u.ID = claims.UserID
	msg.User = &u

	// outbox event for reliable delivery (worker fallback if redis down)
	payload, _ := json.Marshal(map[string]interface{}{
		"message_id": msg.ID,
		"channel_id": msg.ChannelID,
		"project_id": msg.ProjectID,
		"user": u,
		"message": msg.Message,
		"message_type": msg.MessageType,
		"created_at": msg.CreatedAt,
	})
	_, err = tx.Exec(ctx,
		`INSERT INTO outbox_events (organization_id, project_id, event_type, aggregate_type, aggregate_id, payload_json)
		 VALUES ($1,$2,'chat.message.created','chat_message',$3,$4)`,
		orgID, projectID, msg.ID, payload)
	if err != nil {
		log.Printf("chat outbox insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		log.Printf("chat commit: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// publish to redis directly (fast path) — hub will fan-out via pubsub
	if h.Redis != nil {
		redisPayload, _ := json.Marshal(map[string]interface{}{
			"id": msg.ID,
			"type": "chat.message.created",
			"organization_id": orgID,
			"project_id": projectID,
			"channel_id": channelID,
			"occurred_at": msg.CreatedAt,
			"data": msg,
		})
		channel := "chat:" + orgID + ":" + projectID
		if err := h.Redis.Publish(ctx, channel, redisPayload).Err(); err != nil {
			log.Printf("redis publish chat: %v", err)
		}
		// also legacy project channel for tasks compat
	}

	writeJSON(w, http.StatusCreated, msg)
}
