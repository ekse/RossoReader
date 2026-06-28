package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ekse/rssreader/internal/auth"
	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/middleware"
	"github.com/ekse/rssreader/internal/store"
)

const (
	defaultSessionMaxAgeDays = 30
	defaultCookieName        = "session"
)

type AuthHandler struct {
	Store store.Store
}

func NewAuthHandler(s store.Store) *AuthHandler {
	return &AuthHandler{Store: s}
}

func (h *AuthHandler) sessionMaxAge() int {
	if v := os.Getenv("SESSION_MAX_AGE_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultSessionMaxAgeDays
}

func (h *AuthHandler) cookieName() string {
	if v := os.Getenv("SESSION_COOKIE_NAME"); v != "" {
		return v
	}
	return defaultCookieName
}

func (h *AuthHandler) cookieSecure() bool {
	v := os.Getenv("COOKIE_SECURE")
	return v == "true" || v == "1"
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{ID: u.ID, Username: u.Username, IsAdmin: u.IsAdmin}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, hash, err := h.Store.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ok, err := auth.VerifyPassword(hash, req.Password)
	if err != nil || !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionID := uuid.New()
	maxAgeDays := h.sessionMaxAge()
	expiresAt := time.Now().Add(time.Duration(maxAgeDays) * 24 * time.Hour)

	if err := h.Store.CreateSession(r.Context(), sessionID, user.ID, expiresAt); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName(),
		Value:    sessionID.String(),
		Path:     "/",
		MaxAge:   maxAgeDays * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   h.cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookieName())
	if err == nil {
		if id, perr := uuid.Parse(cookie.Value); perr == nil {
			_ = h.Store.DeleteSession(r.Context(), id)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Cleanup expired sessions opportunistically.
	_ = h.Store.DeleteExpiredSessions(r.Context())

	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "current_password and new_password are required", http.StatusBadRequest)
		return
	}

	_, hash, err := h.Store.GetUserByID(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	verified, err := auth.VerifyPassword(hash, req.CurrentPassword)
	if err != nil || !verified {
		http.Error(w, "current password is incorrect", http.StatusUnauthorized)
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.Store.UpdateUserPassword(r.Context(), u.ID, newHash); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := h.Store.CreateUser(r.Context(), req.Username, hash, req.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	caller, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if id == caller.ID {
		http.Error(w, "cannot delete yourself", http.StatusBadRequest)
		return
	}

	count, err := h.Store.CountUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count <= 1 {
		http.Error(w, "cannot delete the last user", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ErrUnauthorized is exported for tests.
var ErrUnauthorized = errors.New("unauthorized")
