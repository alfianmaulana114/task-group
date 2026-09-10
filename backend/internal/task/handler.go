package task

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alfianhamzah/task-group/backend/internal/auth"
	"github.com/alfianhamzah/task-group/backend/internal/org"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func generateTaskKey() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "TG-" + strings.ToUpper(hex.EncodeToString(b))
}

type Handler struct {
	DB *pgxpool.Pool
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority"`
	Status      string `json:"status,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
}

type updateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type taskResponse struct {
	ID          string    `json:"id"`
	TaskKey     string    `json:"task_key"`
	Key         string    `json:"key"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	AssigneeID  string    `json:"assignee_id"`
	ReporterID  string    `json:"reporter_id"`
	CreatedBy   string    `json:"created_by"`
	ProjectID   string    `json:"project_id"`
	DueDate     string    `json:"due_date,omitempty"`
	StartDate   string    `json:"start_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "project_id is required"})
		return
	}

	// Verify project belongs to organization and user is member
	var role string
	err := h.DB.QueryRow(r.Context(),
		`SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
		orgID, claims.UserID,
	).Scan(&role)
	if err != nil {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "you are not a member of this organization"})
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Priority = strings.TrimSpace(strings.ToLower(req.Priority))
	if req.Priority == "" {
		req.Priority = "medium"
	}
	if !isValidPriority(req.Priority) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "priority must be one of: low, medium, high, urgent"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "title is required"})
		return
	}

	req.Status = normalizeStatus(req.Status)
	if req.Status == "" {
		req.Status = "todo"
	}
	if !isValidStatus(req.Status) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "status must be one of: backlog, todo, in_progress, in_review, done, blocked, cancelled"})
		return
	}

	ctx := r.Context()
	// validate assignee is member of org if provided
	var assignee interface{} = nil
	if strings.TrimSpace(req.AssigneeID) != "" {
		var isMember bool
		if err := h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_members WHERE organization_id=$1 AND user_id=$2)`, orgID, req.AssigneeID).Scan(&isMember); err == nil && !isMember {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "assignee is not a member of this organization"})
			return
		}
		assignee = req.AssigneeID
	}
	var t taskResponse
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		taskKey := generateTaskKey()
		var dueDate interface{} = nil
		if strings.TrimSpace(req.DueDate) != "" {
			dueDate = strings.TrimSpace(req.DueDate)
		}
		var startDate interface{} = nil
		if strings.TrimSpace(req.StartDate) != "" {
			startDate = strings.TrimSpace(req.StartDate)
		}
		err = h.DB.QueryRow(ctx,
			`INSERT INTO tasks (task_key, title, description, status, priority, assignee_id, reporter_id, organization_id, project_id, due_date, start_date)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			 RETURNING id, task_key, title, COALESCE(description,''), status, priority, COALESCE(assignee_id::text,''), reporter_id, created_at, project_id, COALESCE(due_date::text,''), COALESCE(start_date::text,'')`,
			taskKey, req.Title, req.Description, req.Status, req.Priority, assignee, claims.UserID, orgID, projectID, dueDate, startDate,
		).Scan(&t.ID, &t.TaskKey, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.CreatedAt, &t.ProjectID, &t.DueDate, &t.StartDate)
		if err == nil {
			t.Status = denormalizeStatus(t.Status)
			t.Key = t.TaskKey
			t.CreatedBy = t.ReporterID

			// Log activity
			go h.logActivity(ctx, orgID, projectID, claims.UserID, "created", "task", t.ID, t.Title, nil, map[string]string{
				"title":    t.Title,
				"status":   t.Status,
				"priority": t.Priority,
			})

			writeJSON(w, http.StatusCreated, t)
			return
		}
		if strings.Contains(err.Error(), "tasks_project_key_unique") || strings.Contains(err.Error(), "tasks_org_project_key_unique") {
			lastErr = err
			continue
		}
		break
	}
	log.Printf("create task: insert: %v last=%v", err, lastErr)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
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

	// Verify project belongs to organization and user is member
	var role string
	err := h.DB.QueryRow(r.Context(),
		`SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
		orgID, claims.UserID,
	).Scan(&role)
	if err != nil {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "you are not a member of this organization"})
		return
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := parsePositiveInt(v); err == nil && n >= 0 {
			offset = n
		}
	}
	rows, err := h.DB.Query(r.Context(),
		`SELECT id, task_key, title, COALESCE(description,''), status, priority, COALESCE(assignee_id::text,''), reporter_id, created_at, project_id, COALESCE(due_date::text,''), COALESCE(start_date::text,'')
		   FROM tasks
		   WHERE organization_id = $1 AND project_id = $2 AND deleted_at IS NULL
		   ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		orgID, projectID, limit, offset,
	)
	if err != nil {
		log.Printf("list tasks: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var tasks []taskResponse
	for rows.Next() {
		var t taskResponse
		if err := rows.Scan(&t.ID, &t.TaskKey, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.CreatedAt, &t.ProjectID, &t.DueDate, &t.StartDate); err != nil {
			log.Printf("list tasks: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		t.Status = denormalizeStatus(t.Status)
		t.Key = t.TaskKey
		t.CreatedBy = t.ReporterID
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []taskResponse{}
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) ListAllTasks(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization required"})
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT t.id, t.task_key, t.title, COALESCE(t.description,''), t.status, t.priority,
		        COALESCE(t.assignee_id::text,''), t.reporter_id, t.created_at, t.project_id,
		        COALESCE(t.due_date::text,''), COALESCE(t.start_date::text,'')
		   FROM tasks t
		   WHERE t.organization_id = $1 AND t.deleted_at IS NULL
		   ORDER BY t.due_date ASC NULLS LAST, t.created_at DESC`,
		orgID,
	)
	if err != nil {
		log.Printf("list all tasks: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var tasks []taskResponse
	for rows.Next() {
		var t taskResponse
		if err := rows.Scan(&t.ID, &t.TaskKey, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.CreatedAt, &t.ProjectID, &t.DueDate, &t.StartDate); err != nil {
			log.Printf("list all tasks: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		t.Status = denormalizeStatus(t.Status)
		t.Key = t.TaskKey
		t.CreatedBy = t.ReporterID
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []taskResponse{}
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	taskID := chi.URLParam(r, "task_id")
	if orgID == "" || taskID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and task ID required"})
		return
	}

	// Verify task belongs to organization (project_id derived from task)
	var t taskResponse
	err := h.DB.QueryRow(r.Context(),
		`SELECT id, task_key, title, COALESCE(description,''), status, priority, COALESCE(assignee_id::text,''), reporter_id, created_at, project_id, COALESCE(due_date::text,''), COALESCE(start_date::text,'')
		   FROM tasks WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		taskID, orgID,
	).Scan(&t.ID, &t.TaskKey, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.CreatedAt, &t.ProjectID, &t.DueDate, &t.StartDate)
	if err == nil {
		t.Status = denormalizeStatus(t.Status)
		t.Key = t.TaskKey
		t.CreatedBy = t.ReporterID
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "task not found"})
			return
		}
		log.Printf("get task: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	taskID := chi.URLParam(r, "task_id")
	if orgID == "" || taskID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and task ID required"})
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.Status != nil {
		*req.Status = normalizeStatus(*req.Status)
		if !isValidStatus(*req.Status) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "status must be one of: backlog, todo, in_progress, in_review, done, blocked, cancelled"})
			return
		}
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "title cannot be empty"})
			return
		}
		req.Title = &title
	}

	ctx := r.Context()
	var t taskResponse
	// nullable fields: distinguish not-provided (nil) vs clear ("" -> NULL) vs set
	hasAssignee := req.AssigneeID != nil
	var assigneeVal interface{} = nil
	if hasAssignee && strings.TrimSpace(*req.AssigneeID) != "" {
		assigneeVal = strings.TrimSpace(*req.AssigneeID)
		// validate assignee membership
		var isMember bool
		if err := h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_members WHERE organization_id=$1 AND user_id=$2)`, orgID, assigneeVal).Scan(&isMember); err == nil && !isMember {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "assignee is not a member of this organization"})
			return
		}
	} else if hasAssignee {
		assigneeVal = nil // clear
	}
	hasDueDate := req.DueDate != nil
	var dueDateVal interface{} = nil
	if hasDueDate && strings.TrimSpace(*req.DueDate) != "" {
		dueDateVal = strings.TrimSpace(*req.DueDate)
	} else if hasDueDate {
		dueDateVal = nil
	}
	hasStartDate := req.StartDate != nil
	var startDateVal interface{} = nil
	if hasStartDate && strings.TrimSpace(*req.StartDate) != "" {
		startDateVal = strings.TrimSpace(*req.StartDate)
	} else if hasStartDate {
		startDateVal = nil
	}
	err := h.DB.QueryRow(ctx,
		`UPDATE tasks SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			priority = COALESCE($4, priority),
			assignee_id = CASE WHEN $5 THEN $6::uuid ELSE assignee_id END,
			due_date = CASE WHEN $7 THEN $8::date ELSE due_date END,
			start_date = CASE WHEN $9 THEN $10::date ELSE start_date END,
			status = COALESCE($11, status)
		 WHERE id = $1 AND organization_id = $12 AND deleted_at IS NULL
		 RETURNING id, task_key, title, COALESCE(description,''), status, priority, COALESCE(assignee_id::text,''), reporter_id, created_at, project_id, COALESCE(due_date::text,''), COALESCE(start_date::text,'')`,
		taskID, req.Title, req.Description, req.Priority, hasAssignee, assigneeVal, hasDueDate, dueDateVal, hasStartDate, startDateVal, req.Status, orgID,
	).Scan(&t.ID, &t.TaskKey, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.CreatedAt, &t.ProjectID, &t.DueDate, &t.StartDate)
	if err == nil {
		t.Status = denormalizeStatus(t.Status)
		t.Key = t.TaskKey
		t.CreatedBy = t.ReporterID
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "task not found"})
			return
		}
		log.Printf("update task: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Log activity
	go h.logActivity(ctx, orgID, "", claims.UserID, "updated", "task", t.ID, t.Title, nil, map[string]string{
		"title":    t.Title,
		"status":   t.Status,
		"priority": t.Priority,
	})

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	taskID := chi.URLParam(r, "task_id")
	if orgID == "" || taskID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and task ID required"})
		return
	}

	ctx := r.Context()

	// Get task title before deletion
	var taskTitle string
	_ = h.DB.QueryRow(ctx, `SELECT title FROM tasks WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`, taskID, orgID).Scan(&taskTitle)

	_, err := h.DB.Exec(ctx,
		`UPDATE tasks SET deleted_at = now() WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		taskID, orgID,
	)
	if err != nil {
		log.Printf("delete task: exec: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Log activity
	go h.logActivity(ctx, orgID, "", claims.UserID, "deleted", "task", taskID, taskTitle, nil, nil)

	writeJSON(w, http.StatusOK, map[string]string{"status": "task deleted"})
}

func isValidPriority(p string) bool {
	valid := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	return valid[strings.ToLower(p)]
}

func normalizeStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "in_review" {
		return "review"
	}
	return s
}

func denormalizeStatus(s string) string {
	if s == "review" {
		return "in_review"
	}
	return s
}

func isValidStatus(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	valid := map[string]bool{
		"backlog":     true,
		"todo":        true,
		"in_progress": true,
		"review":      true,
		"in_review":   true,
		"done":        true,
		"blocked":     true,
		"cancelled":   true,
	}
	return valid[s]
}

func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not numeric")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

type createCommentRequest struct {
	Body string `json:"body"`
}

type commentResponse struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"task_id"`
	UserID    string     `json:"user_id"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	UserName  string     `json:"user_name"`
	UserEmail string     `json:"user_email"`
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	taskID := chi.URLParam(r, "task_id")
	if orgID == "" || taskID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and task ID required"})
		return
	}

	ctx := r.Context()
	rows, err := h.DB.Query(ctx,
		`SELECT c.id, c.task_id, c.user_id, c.body, c.created_at, c.edited_at,
		        COALESCE(u.full_name, '') as user_name, COALESCE(u.email, '') as user_email
		 FROM comments c
		 LEFT JOIN users u ON u.id = c.user_id
		 WHERE c.task_id = $1 AND c.organization_id = $2 AND c.deleted_at IS NULL
		 ORDER BY c.created_at ASC`,
		taskID, orgID,
	)
	if err != nil {
		log.Printf("list comments: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var comments []commentResponse
	for rows.Next() {
		var c commentResponse
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt, &c.EditedAt, &c.UserName, &c.UserEmail); err != nil {
			log.Printf("list comments: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		comments = append(comments, c)
	}

	if comments == nil {
		comments = []commentResponse{}
	}

	writeJSON(w, http.StatusOK, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	taskID := chi.URLParam(r, "task_id")
	if orgID == "" || taskID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and task ID required"})
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	body := strings.TrimSpace(req.Body)
	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "body tidak boleh kosong"})
		return
	}
	if len(body) > 5000 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "body maksimal 5000 karakter"})
		return
	}

	ctx := r.Context()

	// Verify task exists and belongs to org
	var exists bool
	err := h.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL)`,
		taskID, orgID,
	).Scan(&exists)
	if err != nil || !exists {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "task not found"})
		return
	}

	var c commentResponse
	err = h.DB.QueryRow(ctx,
		`INSERT INTO comments (task_id, organization_id, project_id, user_id, body)
		 VALUES ($1, $2, (SELECT project_id FROM tasks WHERE id = $1), $3, $4)
		 RETURNING id, task_id, user_id, body, created_at, NULL`,
		taskID, orgID, claims.UserID, body,
	).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Body, &c.CreatedAt, &c.EditedAt)
	if err != nil {
		log.Printf("create comment: insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Fetch user name
	_ = h.DB.QueryRow(ctx, `SELECT COALESCE(full_name, ''), COALESCE(email, '') FROM users WHERE id = $1`, claims.UserID).Scan(&c.UserName, &c.UserEmail)

	// Log activity
	go h.logActivity(ctx, orgID, "", claims.UserID, "created", "comment", c.ID, body, nil, map[string]string{"body": body})

	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	commentID := chi.URLParam(r, "comment_id")
	if orgID == "" || commentID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and comment ID required"})
		return
	}

	ctx := r.Context()

	// Only allow deleting own comments or if user is admin/owner
	var commentUserID, commenterRole string
	err := h.DB.QueryRow(ctx,
		`SELECT c.user_id, COALESCE(om.role, 'member')
		 FROM comments c
		 LEFT JOIN organization_members om ON om.organization_id = c.organization_id AND om.user_id = $1
		 WHERE c.id = $2 AND c.organization_id = $3 AND c.deleted_at IS NULL`,
		claims.UserID, commentID, orgID,
	).Scan(&commentUserID, &commenterRole)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "comment not found"})
		return
	}

	if commentUserID != claims.UserID && commenterRole != "owner" && commenterRole != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "cannot delete other user's comment"})
		return
	}

	_, err = h.DB.Exec(ctx,
		`UPDATE comments SET deleted_at = now() WHERE id = $1 AND organization_id = $2`,
		commentID, orgID,
	)
	if err != nil {
		log.Printf("delete comment: exec: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Log activity
	go h.logActivity(ctx, orgID, "", claims.UserID, "deleted", "comment", commentID, "", nil, nil)

	writeJSON(w, http.StatusOK, map[string]string{"status": "comment deleted"})
}

func (h *Handler) logActivity(ctx context.Context, orgID, projectID, actorUserID, actionType, entityType, entityID, entityTitle string, before, after interface{}) {
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
