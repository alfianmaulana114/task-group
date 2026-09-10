package org

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alfianhamzah/task-group/backend/internal/auth"
)

type Handler struct {
	DB *pgxpool.Pool
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type updateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type projectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type updateOrgRequest struct {
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type updateMemberRoleRequest struct {
	Role string `json:"role"`
}

type orgResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OwnerUserID string    `json:"owner_user_id"`
	Plan        string    `json:"plan"`
	Status      string    `json:"status"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type memberResponse struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	Role     string     `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
	User     *userBrief `json:"user,omitempty"`
}

type userBrief struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT o.id, o.name, o.slug, o.owner_user_id, o.plan, o.status, om.role, o.created_at
		 FROM organizations o
		 JOIN organization_members om ON om.organization_id = o.id
		 WHERE om.user_id = $1 AND o.deleted_at IS NULL
		 ORDER BY o.created_at DESC`,
		claims.UserID,
	)
	if err != nil {
		log.Printf("list orgs: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var orgs []orgResponse
	for rows.Next() {
		var o orgResponse
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerUserID, &o.Plan, &o.Status, &o.Role, &o.CreatedAt); err != nil {
			log.Printf("list orgs: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		orgs = append(orgs, o)
	}

	if orgs == nil {
		orgs = []orgResponse{}
	}

	writeJSON(w, http.StatusOK, orgs)
}

func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))

	if req.Name == "" || req.Slug == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name and slug are required"})
		return
	}

	if len(req.Slug) < 3 || len(req.Slug) > 50 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "slug must be between 3 and 50 characters"})
		return
	}

	for _, c := range req.Slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "slug must contain only lowercase letters, numbers, and hyphens"})
			return
		}
	}

	ctx := r.Context()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		log.Printf("create org: begin tx: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var org orgResponse
	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (name, slug, owner_user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, slug, owner_user_id, plan, status, created_at`,
		req.Name, req.Slug, claims.UserID,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerUserID, &org.Plan, &org.Status, &org.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "organizations_slug_unique") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "slug already taken"})
			return
		}
		log.Printf("create org: insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO organization_members (organization_id, user_id, role)
		 VALUES ($1, $2, 'owner')`,
		org.ID, claims.UserID,
	)
	if err != nil {
		log.Printf("create org: insert owner membership: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		log.Printf("create org: commit: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, org)
}

func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	var o orgResponse
	err := h.DB.QueryRow(r.Context(),
		`SELECT id, name, slug, owner_user_id, plan, status, created_at
		 FROM organizations WHERE id = $1 AND deleted_at IS NULL`,
		orgID,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerUserID, &o.Plan, &o.Status, &o.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "organization not found"})
			return
		}
		log.Printf("get org: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	role := GetOrgRole(r.Context())

	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can update organization"})
		return
	}

	var req updateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == nil && req.Slug == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "at least one field (name or slug) is required"})
		return
	}

	ctx := r.Context()

	if req.Slug != nil {
		slug := strings.TrimSpace(strings.ToLower(*req.Slug))
		if len(slug) < 3 || len(slug) > 50 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "slug must be between 3 and 50 characters"})
			return
		}
		for _, c := range slug {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "slug must contain only lowercase letters, numbers, and hyphens"})
				return
			}
		}
		req.Slug = &slug
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
	}

	var o orgResponse
	err := h.DB.QueryRow(ctx,
		`UPDATE organizations SET
			name = COALESCE($2, name),
			slug = COALESCE($3, slug)
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING id, name, slug, owner_user_id, plan, status, created_at`,
		orgID, req.Name, req.Slug,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerUserID, &o.Plan, &o.Status, &o.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "organizations_slug_unique") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "slug already taken"})
			return
		}
		log.Printf("update org: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, o)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// keep alnum and hyphen only
	var out strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			prevDash = false
		} else if r == '-' {
			if !prevDash {
				out.WriteRune(r)
				prevDash = true
			}
		}
	}
	res := strings.Trim(out.String(), "-")
	if res == "" {
		res = "project"
	}
	return res
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}

	ctx := r.Context()
	// Check user is member of organization (owner or admin or member)
	var role string
	err := h.DB.QueryRow(ctx,
		`SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
		orgID, claims.UserID,
	).Scan(&role)
	if err != nil {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "you are not a member of this organization"})
		return
	}

	// Only owner/admin can create projects
	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can create projects"})
		return
	}

	slug := slugify(req.Name)

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		log.Printf("create project: begin tx: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var p projectResponse
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (name, slug, description, organization_id, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (organization_id, slug) DO NOTHING
		 RETURNING id, name, COALESCE(description,''), slug, status, created_by, created_at`,
		req.Name, slug, req.Description, orgID, claims.UserID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Slug, &p.Status, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "project slug already taken"})
			return
		}
		if strings.Contains(err.Error(), "projects_org_slug_unique") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "project slug already taken"})
			return
		}
		log.Printf("create project: insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, 'manager') ON CONFLICT DO NOTHING`,
		p.ID, claims.UserID)
	if err != nil {
		log.Printf("create project: insert project member: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("create project: commit: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	orgID := GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n := parsePositiveInt(v); n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n := parsePositiveInt(v); n >= 0 {
			offset = n
		}
	}
	rows, err := h.DB.Query(r.Context(),
		`SELECT id, name, COALESCE(description,''), slug, status, created_by, created_at
		   FROM projects
		   WHERE organization_id = $1 AND deleted_at IS NULL
		   ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		orgID, limit, offset,
	)
	if err != nil {
		log.Printf("list projects: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var projects []projectResponse
	for rows.Next() {
		var p projectResponse
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Slug, &p.Status, &p.CreatedBy, &p.CreatedAt); err != nil {
			log.Printf("list projects: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []projectResponse{}
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	projectID := chi.URLParam(r, "project_id")
	if orgID == "" || projectID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization and project ID required"})
		return
	}

	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	// Verify project belongs to organization
	var role string
	err := h.DB.QueryRow(r.Context(),
		`SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
		orgID, claims.UserID,
	).Scan(&role)
	if err != nil {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "you are not a member of this organization"})
		return
	}

	var p projectResponse
	err = h.DB.QueryRow(r.Context(),
		`SELECT id, name, COALESCE(description,''), slug, status, created_by, created_at
		   FROM projects WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		projectID, orgID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Slug, &p.Status, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "project not found"})
			return
		}
		log.Printf("get project: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	role := GetOrgRole(r.Context())

	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can update project"})
		return
	}

	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "project_id is required"})
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	ctx := r.Context()

	var p projectResponse
	err := h.DB.QueryRow(ctx,
		`UPDATE projects SET
			name = COALESCE($2, name),
			description = COALESCE($3, description)
		 WHERE id = $1 AND organization_id = $4 AND deleted_at IS NULL
		 RETURNING id, name, COALESCE(description,''), slug, status, created_by, created_at`,
		projectID, req.Name, req.Description, orgID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Slug, &p.Status, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "project not found"})
			return
		}
		if strings.Contains(err.Error(), "projects_org_slug_unique") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "project slug already taken"})
			return
		}
		log.Printf("update project: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	if orgID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "organization context required"})
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT om.id, om.user_id, om.role, om.joined_at,
		        u.id, u.email, u.full_name
		 FROM organization_members om
		 JOIN users u ON u.id = om.user_id
		 WHERE om.organization_id = $1 AND u.deleted_at IS NULL
		 ORDER BY om.joined_at ASC LIMIT 200`,
		orgID,
	)
	if err != nil {
		log.Printf("list members: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer rows.Close()

	var members []memberResponse
	for rows.Next() {
		var m memberResponse
		var u userBrief
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.JoinedAt, &u.ID, &u.Email, &u.FullName); err != nil {
			log.Printf("list members: scan: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
		m.User = &u
		members = append(members, m)
	}

	if members == nil {
		members = []memberResponse{}
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	role := GetOrgRole(r.Context())

	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can add members"})
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))

	if req.UserID == "" && req.Email == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email is required"})
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if req.Role != "admin" && req.Role != "member" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "role must be admin or member"})
		return
	}

	ctx := r.Context()

	// Resolve user_id from email if not provided
	if req.UserID == "" && req.Email != "" {
		err := h.DB.QueryRow(ctx,
			`SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`,
			req.Email,
		).Scan(&req.UserID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found with this email"})
			return
		}
	}

	var exists bool
	err := h.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`,
		req.UserID,
	).Scan(&exists)
	if err != nil || !exists {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		return
	}

	// protect owner from being downgraded
	var existingRole string
	err = h.DB.QueryRow(ctx, `SELECT role FROM organization_members WHERE organization_id=$1 AND user_id=$2`, orgID, req.UserID).Scan(&existingRole)
	if err == nil && existingRole == "owner" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot change owner role"})
		return
	}
	var m memberResponse
	err = h.DB.QueryRow(ctx,
		`INSERT INTO organization_members (organization_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role WHERE organization_members.role != 'owner'
		 RETURNING id, user_id, role, joined_at`,
		orgID, req.UserID, req.Role,
	).Scan(&m.ID, &m.UserID, &m.Role, &m.JoinedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot change owner role"})
			return
		}
		log.Printf("add member: insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, m)
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	role := GetOrgRole(r.Context())

	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can update member roles"})
		return
	}

	memberID := chi.URLParam(r, "member_id")
	if memberID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "member_id is required"})
		return
	}

	var req updateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	if req.Role != "admin" && req.Role != "member" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "role must be admin or member"})
		return
	}

	ctx := r.Context()

	var targetRole string
	err := h.DB.QueryRow(ctx,
		`SELECT role FROM organization_members WHERE id = $1 AND organization_id = $2`,
		memberID, orgID,
	).Scan(&targetRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "member not found"})
			return
		}
		log.Printf("update member role: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	if targetRole == "owner" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot change owner role"})
		return
	}

	var m memberResponse
	err = h.DB.QueryRow(ctx,
		`UPDATE organization_members SET role = $1
		 WHERE id = $2 AND organization_id = $3
		 RETURNING id, user_id, role, joined_at`,
		req.Role, memberID, orgID,
	).Scan(&m.ID, &m.UserID, &m.Role, &m.JoinedAt)
	if err != nil {
		log.Printf("update member role: update: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID := GetOrgID(r.Context())
	role := GetOrgRole(r.Context())

	if role != "owner" && role != "admin" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "only owner or admin can remove members"})
		return
	}

	memberID := chi.URLParam(r, "member_id")
	if memberID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "member_id is required"})
		return
	}

	ctx := r.Context()

	var targetRole string
	err := h.DB.QueryRow(ctx,
		`SELECT role FROM organization_members WHERE id = $1 AND organization_id = $2`,
		memberID, orgID,
	).Scan(&targetRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "member not found"})
			return
		}
		log.Printf("remove member: query: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	if targetRole == "owner" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "cannot remove the owner"})
		return
	}

	_, err = h.DB.Exec(ctx,
		`DELETE FROM organization_members WHERE id = $1 AND organization_id = $2`,
		memberID, orgID,
	)
	if err != nil {
		log.Printf("remove member: delete: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "member removed"})
}

func parsePositiveInt(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
