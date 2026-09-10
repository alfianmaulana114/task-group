package search

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/alfianhamzah/task-group/backend/internal/auth"
	"github.com/alfianhamzah/task-group/backend/internal/org"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

type searchResult struct {
	Type        string    `json:"type"`
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Slug        string    `json:"slug,omitempty"`
	Status      string    `json:"status,omitempty"`
	Priority    string    `json:"priority,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsKey).(*auth.Claims)
	if !ok || claims == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	orgID := org.GetOrgID(r.Context())
	if orgID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "organization required"})
		return
	}

	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]searchResult{})
		return
	}

	tsQuery := toTsQuery(q)

	ctx := r.Context()
	var results []searchResult

	// Search tasks
	rows, err := h.DB.Query(ctx,
		`SELECT t.id, t.title, COALESCE(t.description,''), t.status, t.priority,
		        t.project_id, COALESCE(p.name,''), t.created_at
		   FROM tasks t
		   LEFT JOIN projects p ON p.id = t.project_id
		   WHERE t.organization_id = $1 AND t.deleted_at IS NULL
		     AND to_tsvector('simple', COALESCE(t.title,'') || ' ' || COALESCE(t.description,'') || ' ' || COALESCE(t.task_key,'')) @@ to_tsquery('simple', $2)
		   ORDER BY ts_rank(to_tsvector('simple', COALESCE(t.title,'') || ' ' || COALESCE(t.description,'') || ' ' || COALESCE(t.task_key,'')), to_tsquery('simple', $2)) DESC
		   LIMIT 10`,
		orgID, tsQuery,
	)
	if err != nil {
		log.Printf("search tasks: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var s searchResult
			if err := rows.Scan(&s.ID, &s.Title, &s.Subtitle, &s.Status, &s.Priority, &s.ProjectID, &s.ProjectName, &s.CreatedAt); err == nil {
				s.Type = "task"
				results = append(results, s)
			}
		}
	}

	// Search projects
	rows2, err := h.DB.Query(ctx,
		`SELECT id, name, slug, COALESCE(description,''), created_at
		   FROM projects
		   WHERE organization_id = $1 AND deleted_at IS NULL
		     AND to_tsvector('simple', COALESCE(name,'') || ' ' || COALESCE(description,'')) @@ to_tsquery('simple', $2)
		   ORDER BY ts_rank(to_tsvector('simple', COALESCE(name,'') || ' ' || COALESCE(description,'')), to_tsquery('simple', $2)) DESC
		   LIMIT 5`,
		orgID, tsQuery,
	)
	if err != nil {
		log.Printf("search projects: %v", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var s searchResult
			if err := rows2.Scan(&s.ID, &s.Title, &s.Slug, &s.Subtitle, &s.CreatedAt); err == nil {
				s.Type = "project"
				results = append(results, s)
			}
		}
	}

	// Search comments
	rows3, err := h.DB.Query(ctx,
		`SELECT c.id, LEFT(c.body, 100), t.id, t.title, c.created_at
		   FROM comments c
		   JOIN tasks t ON t.id = c.task_id
		   WHERE t.organization_id = $1
		     AND to_tsvector('simple', c.body) @@ to_tsquery('simple', $2)
		   ORDER BY c.created_at DESC
		   LIMIT 5`,
		orgID, tsQuery,
	)
	if err != nil {
		log.Printf("search comments: %v", err)
	} else {
		defer rows3.Close()
		for rows3.Next() {
			var s searchResult
			if err := rows3.Scan(&s.ID, &s.Subtitle, &s.ProjectID, &s.Title, &s.CreatedAt); err == nil {
				s.Type = "comment"
				results = append(results, s)
			}
		}
	}

	if results == nil {
		results = []searchResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func toTsQuery(q string) string {
	// Split on spaces and join with &
	words := splitWords(q)
	if len(words) == 0 {
		return ""
	}
	result := words[0]
	for _, w := range words[1:] {
		result += " & " + w
	}
	return result
}

func splitWords(q string) []string {
	var words []string
	current := ""
	for _, c := range q {
		if c == ' ' || c == '\t' || c == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
