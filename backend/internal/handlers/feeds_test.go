package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/middleware"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

// authReq builds an HTTP request with the user injected in the context.
func authReq(method, target, body string, user domain.User) *http.Request {
	var reader = strings.NewReader("")
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	return req.WithContext(middleware.SetUserInContext(req.Context(), user))
}

// authedRouter returns a router mirroring MountRouter but using a context-inject
// middleware instead of cookie-based authentication, so handlers receive the user.
func authedRouter(h *handlers.Handler, user domain.User) chi.Router {
	r := chi.NewRouter()
	r.Get("/api/health", h.Health)
	r.Post("/api/auth/login", h.Auth.Login)
	r.Post("/api/auth/logout", h.Auth.Logout)

	r.Post("/api/auth/passkey/login/begin", h.Passkey.LoginBegin)
	r.Post("/api/auth/passkey/login/finish", h.Passkey.LoginFinish)

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(middleware.SetUserInContext(r.Context(), user)))
			})
		})
		r.Get("/api/auth/me", h.Auth.Me)
		r.Put("/api/auth/password", h.Auth.ChangePassword)
		r.Post("/api/auth/passkey/register/begin", h.Passkey.RegisterBegin)
		r.Post("/api/auth/passkey/register/finish", h.Passkey.RegisterFinish)
		r.Get("/api/auth/passkeys", h.Passkey.ListPasskeys)
		r.Delete("/api/auth/passkeys/{id}", h.Passkey.DeletePasskey)
		r.Get("/api/feeds", h.ListFeeds)
		r.Post("/api/feeds", h.AddFeed)
		r.Post("/api/feeds/discover", h.DiscoverFeeds)
		r.Delete("/api/feeds/{id}", h.RemoveFeed)
		r.Post("/api/feeds/{id}/refresh", h.RefreshFeed)
		r.Post("/api/feeds/{id}/read-all", h.MarkFeedRead)
		r.Get("/api/items", h.ListItems)
		r.Post("/api/items/read-all", h.MarkAllRead)
		r.Patch("/api/items/{id}", h.UpdateItem)
		r.Get("/api/settings", h.GetSettings)
		r.Patch("/api/settings", h.UpdateSettings)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/api/users", h.Auth.ListUsers)
			r.Post("/api/users", h.Auth.CreateUser)
			r.Delete("/api/users/{id}", h.Auth.DeleteUser)
		})
	})

	return r
}

func makeUser(t *testing.T, store *mockstore.MockStore, username string, isAdmin bool) domain.User {
	t.Helper()
	u, err := store.CreateUser(context.Background(), username, "hash", isAdmin)
	require.NoError(t, err)
	return u
}

func TestListFeeds(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"},
		{ID: 2, UserID: user.ID, URL: "https://b.com/rss", Title: "B"},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var feeds []domain.Feed
	err := json.Unmarshal(w.Body.Bytes(), &feeds)
	require.NoError(t, err)
	assert.Len(t, feeds, 2)
}

func TestListFeeds_Empty(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var feeds []domain.Feed
	json.Unmarshal(w.Body.Bytes(), &feeds)
	assert.Empty(t, feeds)
}

func TestAddFeed(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds", `{"url":"https://example.com/rss"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var feed domain.Feed
	err := json.Unmarshal(w.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/rss", feed.URL)
	assert.Equal(t, user.ID, feed.UserID)
}

func TestAddFeed_EmptyURL(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds", `{"url":""}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveFeed(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("DELETE", "/api/feeds/1", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRemoveFeed_InvalidID(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("DELETE", "/api/feeds/abc", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkFeedRead(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post 1", URL: "https://ex.com/1"},
		{ID: 2, FeedID: 1, GUID: "2", Title: "Post 2", URL: "https://ex.com/2"},
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds/1/read-all", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	items, _, _ := store.GetItems(context.Background(), domain.ItemsQuery{UserID: user.ID, Page: 1, PerPage: 10})
	for _, it := range items {
		assert.True(t, it.Read, "item %d should be read", it.ID)
	}
}

func TestMarkFeedRead_InvalidID(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds/abc/read-all", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

var _ = time.Now
