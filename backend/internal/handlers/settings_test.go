package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func TestGetSettings(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.UpsertSetting(nil, user.ID, "fetch_interval", "30")

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/settings", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settings map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &settings)
	require.NoError(t, err)
	assert.Equal(t, "30", settings["fetch_interval"])
}

func TestUpdateSettings(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PATCH", "/api/settings", `{"fetch_interval":"60"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settings map[string]string
	json.Unmarshal(w.Body.Bytes(), &settings)
	assert.Equal(t, "60", settings["fetch_interval"])
}

func TestGetAdminSettings(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("GET", "/api/admin/settings", "", admin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]int
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 150, resp["items_limit"])
	assert.Equal(t, 200, resp["feeds_limit"])
}

func TestUpdateAdminSettings(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("PATCH", "/api/admin/settings", `{"items_limit":200,"feeds_limit":300}`, admin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]int
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 200, resp["items_limit"])
	assert.Equal(t, 300, resp["feeds_limit"])
}

func TestUpdateAdminSettings_NonAdmin(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "user", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PATCH", "/api/admin/settings", `{"items_limit":200}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateAdminSettings_InvalidItemsLimit(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("PATCH", "/api/admin/settings", `{"items_limit":0}`, admin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateAdminSettings_InvalidFeedsLimit(t *testing.T) {
	store := mockstore.New()
	admin := makeUser(t, store, "admin", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, admin)

	req := authReq("PATCH", "/api/admin/settings", `{"feeds_limit":0}`, admin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
