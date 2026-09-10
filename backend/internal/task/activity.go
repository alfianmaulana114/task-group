package task

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alfianhamzah/task-group/backend/internal/auth"
	"github.com/alfianhamzah/task-group/backend/internal/org"
)

type ActivityHandler struct {
	DB *pgxpool.Pool
}

type activityLogResponse struct {
	ID           string    `json:"id"`
	ActorUserID  string    `json:"actor_user_id"`
	ActorName    string    `json:"actor_name"`
	ActionType   string    `json:"action_type"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	EntityTitle  string    `json:"entity_title"`
	BeforeJSON   *string   `json:"before_json,omitempty"`
	AfterJSON    *string   `json:"after_json,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func getActivityClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(auth.ClaimsKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

func (h *ActivityHandler) LogActivity(ctx context.Context, orgID, projectID, actorUserID, actionType, entityType, entityID, entityTitle string, before, after interface{}) {
	var beforeJSON, afterJSON *string

	if before != nil {
		b, err := json.Marshal(before)
		if err == nil {
			s := string(b)
			beforeJSON = &s
		}
	}
	if after != nil {
		a, err := json.Marshal(after)
		if err == nil {
			s := string(a)
			afterJSON = &s
		}
	}

	_, err := h.DB.Exec(ctx,
		`INSERT INTO activity_logs (organization_id, project_id, actor_user_id, entity_type, entity_id, action_type, before_json, after_json)
		 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8)`,
		orgID, projectID, actorUserID, entityType, entityID, actionType, beforeJSON, afterJSON,
	)
	if err != nil {
		log.Printf("log activity: exec: %v", err)
	}
}

func (h *ActivityHandler) ListActivityLogs(w http.ResponseWriter, r *http.Request) {
	claims := getActivityClaims(r.Context())
	if claims == nil {
		writeActivityJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	if orgID == "" {
		writeActivityJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	// Optional query param: entity_type, entity_id, project_id
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	projectID := r.URL.Query().Get("project_id")
	limit := 50

	ctx := r.Context()
	rows, err := h.DB.Query(ctx,
		`SELECT al.id, al.actor_user_id,
		        COALESCE(u.full_name, u.email, 'Unknown') as actor_name,
		        al.action_type, al.entity_type, al.entity_id,
		        COALESCE(
		            CASE
		                WHEN al.entity_type = 'task' THEN (SELECT title FROM tasks WHERE id = al.entity_id)
		                WHEN al.entity_type = 'project' THEN (SELECT name FROM projects WHERE id = al.entity_id)
		                WHEN al.entity_type = 'comment' THEN (SELECT left(body, 100) FROM comments WHERE id = al.entity_id)
		                ELSE ''
		            END, ''
		        ) as entity_title,
		        al.before_json, al.after_json, al.created_at
		 FROM activity_logs al
		 LEFT JOIN users u ON u.id = al.actor_user_id
		 WHERE al.organization_id = $1
		   AND ($2::text = '' OR al.entity_type = $2)
		   AND ($3::text = '' OR al.entity_id = $3)
		   AND ($4::text = '' OR al.project_id = $4)
		 ORDER BY al.created_at DESC
		 LIMIT $5`,
		orgID, entityType, entityID, projectID, limit,
	)
	if err != nil {
		log.Printf("list activity logs: query: %v", err)
		writeActivityJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var logs []activityLogResponse
	for rows.Next() {
		var l activityLogResponse
		if err := rows.Scan(&l.ID, &l.ActorUserID, &l.ActorName, &l.ActionType, &l.EntityType, &l.EntityID, &l.EntityTitle, &l.BeforeJSON, &l.AfterJSON, &l.CreatedAt); err != nil {
			log.Printf("list activity logs: scan: %v", err)
			writeActivityJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []activityLogResponse{}
	}

	writeActivityJSON(w, http.StatusOK, logs)
}

func writeActivityJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
