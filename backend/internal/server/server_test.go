package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/USERNAME-GITHUB-KAMU/task-group/backend/internal/config"
	"github.com/USERNAME-GITHUB-KAMU/task-group/backend/internal/platform"
)

func TestHealthz(t *testing.T) {
	cfg := config.Config{
		HTTPPort:          "8080",
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	srv := New(cfg, &platform.Dependencies{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
