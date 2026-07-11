package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ekse/rossoreader/internal/domain"
	"github.com/ekse/rossoreader/internal/middleware"
	"github.com/ekse/rossoreader/internal/store/mockstore"
)

func TestAuthenticate_NoCookie_Unauthorized(t *testing.T) {
	store := mockstore.New()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(store, "session")(inner)

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_BadCookie_Unauthorized(t *testing.T) {
	store := mockstore.New()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(store, "session")(inner)

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "not-a-uuid"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_ExpiredSession_Unauthorized(t *testing.T) {
	store := mockstore.New()
	user := domain.User{ID: 1, Username: "alice", IsAdmin: true}
	store.Users = append(store.Users, user)

	sessionUUID := uuid.MustParse("01020304-0506-4708-890a-0b0c0d0e0f10")
	id := sessionUUID
	_ = store.CreateSession(context.Background(), id, user.ID, time.Now().Add(-time.Hour))

	hit := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(store, "session")(inner)

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionUUID.String()})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, hit, "inner handler should not be called for expired session")
}

func TestAuthenticate_ValidSession_SetsUserInContext(t *testing.T) {
	store := mockstore.New()
	user := domain.User{ID: 2, Username: "alice", IsAdmin: true}
	store.Users = append(store.Users, user)
	store.Passwords[user.ID] = "hash"

	sessionUUID := uuid.MustParse("01020304-0506-4708-890a-0b0c0d0e0f10")
	id := sessionUUID
	_ = store.CreateSession(context.Background(), id, user.ID, time.Now().Add(24*time.Hour))

	hit := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.UserFromContext(r.Context())
		assert.True(t, ok, "user should be in context")
		assert.Equal(t, user.Username, u.Username)
		hit = true
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.Authenticate(store, "session")(inner)

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionUUID.String()})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, hit)
}

func TestRequireAdmin_Admin_Passes(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequireAdmin(inner)

	req := httptest.NewRequest("GET", "/api/users", nil)
	req = req.WithContext(middleware.SetUserInContext(req.Context(), domain.User{IsAdmin: true}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_NonAdmin_Forbidden(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequireAdmin(inner)

	req := httptest.NewRequest("GET", "/api/users", nil)
	req = req.WithContext(middleware.SetUserInContext(req.Context(), domain.User{IsAdmin: false}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_NoUser_Forbidden(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RequireAdmin(inner)

	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
