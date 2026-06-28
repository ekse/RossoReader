package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/auth"
	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func makeUserWithPassword(t *testing.T, store *mockstore.MockStore, username, password string, isAdmin bool) (domain.User, string) {
	t.Helper()
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)
	stubU, err := store.CreateUser(context.Background(), username, "stub", isAdmin)
	require.NoError(t, err)
	require.NoError(t, store.UpdateUserPassword(context.Background(), stubU.ID, hash))
	return stubU, hash
}

func TestAuth_Login_Success(t *testing.T) {
	store := mockstore.New()
	user, _ := makeUserWithPassword(t, store, "alice", "passw0rd", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("POST", "/api/auth/login",
		strings.NewReader(`{"username":"alice","password":"passw0rd"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Set-Cookie should be present.
	setCookie := w.Header().Get("Set-Cookie")
	assert.Contains(t, setCookie, "session=")
	assert.Contains(t, setCookie, "HttpOnly")

	var resp struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"is_admin"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "alice", resp.Username)
	assert.True(t, resp.IsAdmin)
}

func TestAuth_Login_BadPassword(t *testing.T) {
	store := mockstore.New()
	user, _ := makeUserWithPassword(t, store, "alice", "passw0rd", false)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("POST", "/api/auth/login",
		strings.NewReader(`{"username":"alice","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_Login_UnknownUser(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("POST", "/api/auth/login",
		strings.NewReader(`{"username":"bob","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_Me(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/auth/me", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"is_admin"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "alice", resp.Username)
}

func TestAuth_ChangePassword(t *testing.T) {
	store := mockstore.New()
	user, _ := makeUserWithPassword(t, store, "alice", "oldpass", false)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PUT", "/api/auth/password",
		`{"current_password":"oldpass","new_password":"newpass"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify new password works by trying to log in.
	req2 := httptest.NewRequest("POST", "/api/auth/login",
		strings.NewReader(`{"username":"alice","password":"newpass"}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestAuth_ChangePassword_WrongCurrent(t *testing.T) {
	store := mockstore.New()
	user, _ := makeUserWithPassword(t, store, "alice", "oldpass", false)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PUT", "/api/auth/password",
		`{"current_password":"wrong","new_password":"newpass"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ListUsers_AdminOnly(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)
	makeUser(t, store, "alice", false)
	makeUser(t, store, "bob", false)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("GET", "/api/users", "", admin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []struct {
		Username string `json:"username"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	assert.Len(t, users, 3) // admin + alice + bob
}

func TestAuth_CreateUser_AdminOnly(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("POST", "/api/users",
		`{"username":"newuser","password":"hunter2","is_admin":false}`, admin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp struct {
		Username string `json:"username"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "newuser", resp.Username)
}

func TestAuth_CreateUser_NonAdmin_Forbidden(t *testing.T) {
	store := mockstore.New()
	alice := makeUser(t, store, "alice", false)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, alice)

	req := authReq("POST", "/api/users",
		`{"username":"newuser","password":"hunter2","is_admin":false}`, alice)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuth_DeleteUser_CannotDeleteSelf(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("DELETE", "/api/users/1", "", admin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuth_DeleteUser_LastUser(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("DELETE", "/api/users/999", "", admin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}