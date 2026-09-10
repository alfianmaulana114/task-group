package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

type Handler struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type authResponse struct {
	User  userResponse `json:"user"`
	Token string       `json:"token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email, password, and full_name are required"})
		return
	}

	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid email format"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	ctx := r.Context()
	var userID string
	err = h.DB.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id`,
		req.Email, string(hash), req.FullName,
	).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_unique") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "email already registered"})
			return
		}
		log.Printf("register: insert user: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	tokens, err := GenerateTokenPair(h.JWTSecret, userID, req.Email)
	if err != nil {
		log.Printf("register: generate token: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	_, err = h.DB.Exec(ctx,
		`INSERT INTO refresh_sessions (user_id, session_token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashToken(tokens.RefreshToken), time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		log.Printf("register: insert refresh session: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)

	writeJSON(w, http.StatusCreated, authResponse{
		User:  userResponse{ID: userID, Email: req.Email, FullName: req.FullName},
		Token: tokens.AccessToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
		return
	}

	ctx := r.Context()
	var id, email, passwordHash, fullName string
	var userStatus string
	err := h.DB.QueryRow(ctx,
		`SELECT id, email, password_hash, full_name, status FROM users WHERE email = $1 AND deleted_at IS NULL`,
		req.Email,
	).Scan(&id, &email, &passwordHash, &fullName, &userStatus)
	if err == nil && userStatus != "" && userStatus != "active" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "akun dinonaktifkan"})
		return
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
			return
		}
		log.Printf("login: query user: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
		return
	}

	tokens, err := GenerateTokenPair(h.JWTSecret, id, email)
	if err != nil {
		log.Printf("login: generate token: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	_, err = h.DB.Exec(ctx,
		`INSERT INTO refresh_sessions (user_id, session_token_hash, expires_at) VALUES ($1, $2, $3)`,
		id, hashToken(tokens.RefreshToken), time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		log.Printf("login: insert refresh session: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)

	writeJSON(w, http.StatusOK, authResponse{
		User:  userResponse{ID: id, Email: email, FullName: fullName},
		Token: tokens.AccessToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil || refreshToken.Value == "" {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing refresh token"})
		return
	}

	ctx := r.Context()
	hashed := hashToken(refreshToken.Value)
	var userID, email, fullName string
	var userStatus string
	err = h.DB.QueryRow(ctx,
		`SELECT u.id, u.email, u.full_name, u.status FROM refresh_sessions rs
		 JOIN users u ON u.id = rs.user_id
		 WHERE (rs.session_token_hash = $1 OR rs.session_token_hash = $2) AND rs.revoked_at IS NULL AND rs.expires_at > now() AND u.deleted_at IS NULL`,
		hashed, refreshToken.Value,
	).Scan(&userID, &email, &fullName, &userStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid refresh token"})
			return
		}
		log.Printf("refresh: query session: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	if userStatus != "active" {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "akun dinonaktifkan"})
		return
	}

	// atomic revoke old + insert new in tx to prevent reuse
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		log.Printf("refresh: begin tx: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()
	ct, err := tx.Exec(ctx, `UPDATE refresh_sessions SET revoked_at = now() WHERE (session_token_hash = $1 OR session_token_hash = $2) AND revoked_at IS NULL`, hashed, refreshToken.Value)
	if err != nil {
		log.Printf("refresh: revoke old session: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	if ct.RowsAffected() == 0 {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "refresh token already used"})
		return
	}

	tokens, err := GenerateTokenPair(h.JWTSecret, userID, email)
	if err != nil {
		log.Printf("refresh: generate token: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_sessions (user_id, session_token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashToken(tokens.RefreshToken), time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		log.Printf("refresh: insert new session: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		log.Printf("refresh: commit: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	setRefreshCookie(w, tokens.RefreshToken)

	writeJSON(w, http.StatusOK, authResponse{
		User:  userResponse{ID: userID, Email: email, FullName: fullName},
		Token: tokens.AccessToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err == nil && refreshToken.Value != "" {
		hashed := hashToken(refreshToken.Value)
		_, _ = h.DB.Exec(r.Context(),
			`UPDATE refresh_sessions SET revoked_at = now() WHERE session_token_hash = $1 OR session_token_hash = $2`,
			hashed, refreshToken.Value,
		)
	}

	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	ctx := r.Context()
	var email, fullName string
	err := h.DB.QueryRow(ctx,
		`SELECT email, full_name FROM users WHERE id = $1 AND deleted_at IS NULL`,
		claims.UserID,
	).Scan(&email, &fullName)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:       claims.UserID,
		Email:    email,
		FullName: fullName,
	})
}

type updateProfileRequest struct {
	FullName *string `json:"full_name,omitempty"`
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.FullName != nil {
		name := strings.TrimSpace(*req.FullName)
		if len(name) < 2 || len(name) > 100 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "nama harus 2-100 karakter"})
			return
		}
	}

	ctx := r.Context()

	if req.FullName != nil {
		_, err := h.DB.Exec(ctx,
			`UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`,
			*req.FullName, claims.UserID,
		)
		if err != nil {
			log.Printf("update profile: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			return
		}
	}

	var email, fullName string
	err := h.DB.QueryRow(ctx,
		`SELECT email, full_name FROM users WHERE id = $1 AND deleted_at IS NULL`,
		claims.UserID,
	).Scan(&email, &fullName)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:       claims.UserID,
		Email:    email,
		FullName: fullName,
	})
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
