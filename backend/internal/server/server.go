package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alfianhamzah/task-group/backend/internal/auth"
	"github.com/alfianhamzah/task-group/backend/internal/chat"
	"github.com/alfianhamzah/task-group/backend/internal/config"
	"github.com/alfianhamzah/task-group/backend/internal/org"
	"github.com/alfianhamzah/task-group/backend/internal/platform"
	"github.com/alfianhamzah/task-group/backend/internal/search"
	"github.com/alfianhamzah/task-group/backend/internal/task"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	times := rl.attempts[key]
	valid := times[:0]
	for _, t := range times {
		if now.Sub(t) <= rl.window {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rl.limit {
		return false
	}
	rl.attempts[key] = append(valid, now)
	return true
}

type Server struct {
	httpServer *http.Server
	deps       *platform.Dependencies
}

func New(cfg config.Config, deps *platform.Dependencies) *Server {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(func(next http.Handler) http.Handler {
		// limit body 1MB to prevent OOM
		return http.MaxBytesHandler(withCORS(next, cfg.CORSAllowedOrigins), 1<<20)
	})

	authHandler := &auth.Handler{
		DB:        deps.Postgres,
		JWTSecret: cfg.JWTSecret,
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if deps == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "error": "dependencies not initialized"})
			return
		}

		if err := deps.Ready(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		loginLimiter := newRateLimiter(10, time.Minute)
		registerLimiter := newRateLimiter(5, time.Minute)

		r.Post("/auth/register", func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			} else if fwd := r.Header.Get("X-Real-Ip"); fwd != "" {
				ip = fwd
			}
			if !registerLimiter.allow("register:" + ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests, please try again later"})
				return
			}
			authHandler.Register(w, r)
		})

		r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			} else if fwd := r.Header.Get("X-Real-Ip"); fwd != "" {
				ip = fwd
			}
			if !loginLimiter.allow("login:" + ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests, please try again later"})
				return
			}
			authHandler.Login(w, r)
		})

		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authHandler.Middleware)
			r.Get("/auth/me", authHandler.Me)
			r.Patch("/auth/me", authHandler.UpdateProfile)

			orgHandler := &org.Handler{DB: deps.Postgres}
			tenantMW := &org.TenantMiddleware{DB: deps.Postgres}

			r.Get("/organizations", orgHandler.ListOrganizations)
			r.Post("/organizations", orgHandler.CreateOrganization)

			r.Route("/organizations/{organization_id}", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)

				r.Get("/", orgHandler.GetOrganization)
				r.Patch("/", orgHandler.UpdateOrganization)

				r.Get("/members", orgHandler.ListMembers)
				r.Post("/members", orgHandler.AddMember)
				r.Patch("/members/{member_id}", orgHandler.UpdateMemberRole)
				r.Delete("/members/{member_id}", orgHandler.RemoveMember)
			})

			// Projects
			projectHandler := &org.Handler{DB: deps.Postgres}
			taskHandler := &task.Handler{DB: deps.Postgres}
			activityHandler := &task.ActivityHandler{DB: deps.Postgres}
			chatHandler := &chat.Handler{DB: deps.Postgres, Redis: deps.Redis}
			chatHub := chat.NewHubWithOrigins(deps.Redis, cfg.CORSAllowedOrigins)

			r.Route("/projects", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)
				r.Get("/", projectHandler.ListProjects)
				r.Post("/", projectHandler.CreateProject)
			})

			r.Route("/projects/{project_id}", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)

				r.Get("/", projectHandler.GetProject)
				r.Patch("/", projectHandler.UpdateProject)

				// Tasks under project
				r.Get("/tasks", taskHandler.ListTasks)
				r.Post("/tasks", taskHandler.CreateTask)

				// Chat under project
				r.Get("/chat", chatHandler.GetMessages)
				r.Post("/chat", chatHandler.PostMessage)
				r.Get("/chat/ws", chatHub.ServeWS)
			})

			// Standalone task routes (get/update/delete by task_id)
			r.Route("/tasks/{task_id}", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)

				r.Get("/", taskHandler.GetTask)
				r.Patch("/", taskHandler.UpdateTask)
				r.Delete("/", taskHandler.DeleteTask)

				// Comments on task
				r.Get("/comments", taskHandler.ListComments)
				r.Post("/comments", taskHandler.CreateComment)
			})

			// All tasks across org (for calendar) — must be after /tasks/{task_id} to avoid route conflict
			r.Route("/my-tasks", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)
				r.Get("/", taskHandler.ListAllTasks)
			})

			// Comment routes (delete by comment_id)
			r.Route("/comments/{comment_id}", func(r chi.Router) {
				r.Use(tenantMW.RequireOrganization)
				r.Delete("/", taskHandler.DeleteComment)
			})

			// Activity logs
			r.Get("/activity-logs", activityHandler.ListActivityLogs)

			// Search
			searchHandler := &search.Handler{DB: deps.Postgres}
			r.Get("/search", searchHandler.Search)
		})
	})

	log.Println("routes registered: auth, organizations, organization members, projects, tasks, chat")

	// Outbox worker: publish pending chat events to Redis (reliable delivery)
	go startOutboxWorker(deps)

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.HTTPPort,
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
		},
		deps: deps,
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func startOutboxWorker(deps *platform.Dependencies) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	tickCount := 0
	for range ticker.C {
		tickCount++
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		tx, err := deps.Postgres.Begin(ctx)
		if err != nil {
			cancel()
			continue
		}
		rows, err := tx.Query(ctx,
			`SELECT id, organization_id, project_id, payload_json, attempt_count FROM outbox_events WHERE status='pending' AND available_at <= now() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 20`)
		if err != nil {
			_ = tx.Rollback(ctx)
			cancel()
			continue
		}
		type evt struct {
			ID      string
			OrgID   string
			ProjID  string
			Payload []byte
			Attempt int
		}
		var events []evt
		for rows.Next() {
			var e evt
			if err := rows.Scan(&e.ID, &e.OrgID, &e.ProjID, &e.Payload, &e.Attempt); err != nil {
				continue
			}
			events = append(events, e)
		}
		rows.Close()
		if err := tx.Commit(ctx); err != nil {
			cancel()
			continue
		}
		for _, e := range events {
			channel := "chat:" + e.OrgID + ":" + e.ProjID
			publishCtx, publishCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := deps.Redis.Publish(publishCtx, channel, e.Payload).Err()
			publishCancel()
			if err != nil {
				log.Printf("outbox redis publish failed id=%s: %v", e.ID, err)
				_, _ = deps.Postgres.Exec(ctx, `UPDATE outbox_events SET attempt_count = attempt_count + 1, available_at = now() + (LEAST(attempt_count, 60) || ' seconds')::interval WHERE id=$1`, e.ID)
				continue
			}
			_, _ = deps.Postgres.Exec(ctx, `UPDATE outbox_events SET status='processed', processed_at=now() WHERE id=$1`, e.ID)
		}
		cancel()

		// Cleanup expired refresh sessions every ~8 minutes (100 ticks * 5s)
		if tickCount%100 == 0 {
			cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, _ = deps.Postgres.Exec(cleanCtx, `DELETE FROM refresh_sessions WHERE expires_at < now() - interval '1 day'`)
			cleanCancel()
		}
	}
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Organization-Id")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
