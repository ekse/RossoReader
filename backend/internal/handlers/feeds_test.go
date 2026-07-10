package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/middleware"
	"github.com/ekse/rssreader/internal/opml"
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
		r.Get("/api/feeds/opml/export", h.ExportOPML)
		r.Post("/api/feeds/opml/preview", h.PreviewOPMLImport)
		r.Get("/api/items", h.ListItems)
		r.Get("/api/items/search", h.SearchItems)
		r.Post("/api/items/read-all", h.MarkAllRead)
		r.Patch("/api/items/{id}", h.UpdateItem)
		r.Get("/api/settings", h.GetSettings)
		r.Patch("/api/settings", h.UpdateSettings)

		r.Get("/api/labels", h.ListLabels)
		r.Post("/api/labels", h.CreateLabel)
		r.Put("/api/labels/{id}", h.UpdateLabel)
		r.Delete("/api/labels/{id}", h.DeleteLabel)

		r.Get("/api/feeds/grouped", h.GroupedFeeds)
		r.Get("/api/feeds/unread-counts", h.UnreadCounts)
		r.Get("/api/feeds/{id}/labels", h.GetFeedLabels)
		r.Post("/api/feeds/{id}/labels", h.AddFeedLabel)
		r.Delete("/api/feeds/{id}/labels/{lid}", h.RemoveFeedLabel)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/api/users", h.Auth.ListUsers)
			r.Post("/api/users", h.Auth.CreateUser)
			r.Delete("/api/users/{id}", h.Auth.DeleteUser)
			r.Get("/api/admin/settings", h.GetAdminSettings)
			r.Patch("/api/admin/settings", h.UpdateAdminSettings)
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

func TestAddFeed_AtLimit(t *testing.T) {
	store := mockstore.New()
	store.SetFeedsLimit(context.TODO(), 1)
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds", `{"url":"https://example.com/rss"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "limit")
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

func TestExportOPML_NoFeeds(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/opml/export", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "<body>")
	assert.Contains(t, w.Body.String(), "</body>")
}

func TestExportOPML_WithFeeds(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "Feed A"},
		{ID: 2, UserID: user.ID, URL: "https://b.com/rss", Title: "Feed B"},
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/opml/export", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

	feeds, err := opml.ParseOPML(strings.NewReader(w.Body.String()))
	require.NoError(t, err)
	require.Len(t, feeds, 2)
	assert.Equal(t, "Feed A", feeds[0].Title)
	assert.Equal(t, "https://a.com/rss", feeds[0].URL)
}

func TestExportOPML_Unauthenticated(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))

	req := httptest.NewRequest("GET", "/api/feeds/opml/export", nil)
	w := httptest.NewRecorder()
	h.MountRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPreviewOPMLImport_ValidFile(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	opmlContent := `<?xml version="1.0"?>
<opml version="1.0">
  <body>
    <outline text="Test Feed" type="rss" xmlUrl="https://test.com/rss"/>
  </body>
</opml>`

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "feeds.opml")
	require.NoError(t, err)
	_, err = io.Copy(fw, strings.NewReader(opmlContent))
	require.NoError(t, err)
	w.Close()

	req := authReq("POST", "/api/feeds/opml/preview", b.String(), user)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var feeds []opml.OpmlFeed
	err = json.Unmarshal(resp.Body.Bytes(), &feeds)
	require.NoError(t, err)
	require.Len(t, feeds, 1)
	assert.Equal(t, "Test Feed", feeds[0].Title)
	assert.Equal(t, "https://test.com/rss", feeds[0].URL)
}

func TestPreviewOPMLImport_InvalidFile(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", "bad.opml")
	require.NoError(t, err)
	_, err = io.Copy(fw, strings.NewReader("not xml"))
	require.NoError(t, err)
	w.Close()

	req := authReq("POST", "/api/feeds/opml/preview", b.String(), user)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPreviewOPMLImport_NoFile(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req := authReq("POST", "/api/feeds/opml/preview", b.String(), user)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPreviewOPMLImport_Unauthenticated(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", "f.opml")
	io.Copy(fw, strings.NewReader("<opml><body></body></opml>"))
	w.Close()

	req := httptest.NewRequest("POST", "/api/feeds/opml/preview", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp := httptest.NewRecorder()
	h.MountRouter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestUnreadCounts(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"},
		{ID: 2, UserID: user.ID, URL: "https://b.com/rss", Title: "B"},
	}
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "a1", Title: "unread"},
		{ID: 2, FeedID: 1, GUID: "a2", Title: "read"},
		{ID: 3, FeedID: 2, GUID: "b1", Title: "unread"},
	}
	// mark item 2 as read for the user
	store.MarkItemRead(context.Background(), user.ID, 2, true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/unread-counts", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]map[int64]int
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp["counts"][1]) // feed 1: 1 unread
	assert.Equal(t, 1, resp["counts"][2]) // feed 2: 1 unread
}

func TestUnreadCounts_AllRead(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"},
	}
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "a1", Title: "read"},
	}
	store.MarkItemRead(context.Background(), user.ID, 1, true)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/unread-counts", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]map[int64]int
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp["counts"][1])
}

func TestUnreadCounts_OtherUser(t *testing.T) {
	store := mockstore.New()
	alice := makeUser(t, store, "alice", true)
	bob := makeUser(t, store, "bob", false)
	// bob owns feed 1 with an unread item
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: bob.ID, URL: "https://bob.com/rss", Title: "Bob"},
	}
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "b1", Title: "unread"},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, alice)

	req := authReq("GET", "/api/feeds/unread-counts", "", alice)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]map[int64]int
	json.Unmarshal(w.Body.Bytes(), &resp)
	_, exists := resp["counts"][1]
	assert.False(t, exists, "alice should not see bob's unread counts")
}
