package org

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/alfianhamzah/task-group/backend/internal/auth"
)

type contextKey string

const (
	orgIDKey  contextKey = "org_id"
	orgRoleKey contextKey = "org_role"
)

type TenantMiddleware struct {
	DB *pgxpool.Pool
}

func (t *TenantMiddleware) RequireOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetClaims(r.Context())
		if claims == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		orgID := chi.URLParam(r, "organization_id")
		if orgID == "" {
			orgID = r.Header.Get("X-Organization-Id")
		}
		if orgID == "" {
			http.Error(w, `{"error":"organization_id is required"}`, http.StatusBadRequest)
			return
		}

		var role string
		err := t.DB.QueryRow(r.Context(),
			`SELECT role FROM organization_members WHERE organization_id = $1 AND user_id = $2`,
			orgID, claims.UserID,
		).Scan(&role)
		if err != nil {
			http.Error(w, `{"error":"you are not a member of this organization"}`, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), orgIDKey, orgID)
		ctx = context.WithValue(ctx, orgRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetOrgID(ctx context.Context) string {
	id, _ := ctx.Value(orgIDKey).(string)
	return id
}

func GetOrgRole(ctx context.Context) string {
	role, _ := ctx.Value(orgRoleKey).(string)
	return role
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetOrgRole(r.Context())
			if role == "" {
				http.Error(w, `{"error":"organization context required"}`, http.StatusForbidden)
				return
			}

			if _, ok := allowed[role]; !ok {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
