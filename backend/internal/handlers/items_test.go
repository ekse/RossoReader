package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
	"github.com/ekse/rssreader/internal/store/pgstore/pgstoretest"
)

func TestListItems(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://x.com/rss", Title: "X"}}
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post 1", URL: "https://ex.com/1", PublishedAt: &now},
		{ID: 2, FeedID: 1, GUID: "2", Title: "Post 2", URL: "https://ex.com/2", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items?per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
		Total int64         `json:"total"`
		Page  int           `json:"page"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, int64(2), resp.Total)
}

func TestListItems_WithFeedFilter(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://x.com/rss", Title: "X"},
		{ID: 2, UserID: user.ID, URL: "https://y.com/rss", Title: "Y"},
	}
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Feed 1 Post", URL: "https://ex.com/1", PublishedAt: &now},
		{ID: 2, FeedID: 2, GUID: "2", Title: "Feed 2 Post", URL: "https://ex.com/2", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items?feed_id=1&per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Items, 1)
	if len(resp.Items) > 0 {
		assert.Equal(t, "Feed 1 Post", resp.Items[0].Title)
	}
}

func TestUpdateItem_MarkRead(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://x.com/rss", Title: "X"}}
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post", URL: "https://ex.com/1", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PATCH", "/api/items/1", `{"read":true}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item domain.Item
	json.Unmarshal(w.Body.Bytes(), &item)
	assert.True(t, item.Read)
}

func TestUpdateItem_MarkStarred(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://x.com/rss", Title: "X"}}
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post", URL: "https://ex.com/1", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PATCH", "/api/items/1", `{"starred":true}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item domain.Item
	json.Unmarshal(w.Body.Bytes(), &item)
	assert.True(t, item.Starred)
}

func TestSearchItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, _, cleanup := pgstoretest.SetupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user, err := store.CreateUser(ctx, "alice", "hash", true)
	require.NoError(t, err)

	feed, err := store.CreateFeed(ctx, user.ID, "https://x.com/rss", "X", "", "", "", "")
	require.NoError(t, err)

	desc := "feed with searchable content about go programming"
	now := time.Now()

	_, err = store.UpsertItem(ctx, domain.Item{
		FeedID:      feed.ID,
		GUID:        "1",
		Title:       "Golang Tips",
		URL:         "https://ex.com/1",
		Description: &desc,
		PublishedAt: &now,
	})
	require.NoError(t, err)

	_, err = store.UpsertItem(ctx, domain.Item{
		FeedID:      feed.ID,
		GUID:        "2",
		Title:       "Rust News",
		URL:         "https://ex.com/2",
		PublishedAt: &now,
	})
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items/search?q=golang&per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
		Total int64         `json:"total"`
		Page  int           `json:"page"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, "Golang Tips", resp.Items[0].Title)
}

func TestSearchItems_EmptyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, _, cleanup := pgstoretest.SetupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user, err := store.CreateUser(ctx, "alice", "hash", true)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items/search?q=", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchItems_WithFeedFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, _, cleanup := pgstoretest.SetupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user, err := store.CreateUser(ctx, "alice", "hash", true)
	require.NoError(t, err)

	feed1, err := store.CreateFeed(ctx, user.ID, "https://x.com/rss", "X", "", "", "", "")
	require.NoError(t, err)

	feed2, err := store.CreateFeed(ctx, user.ID, "https://y.com/rss", "Y", "", "", "", "")
	require.NoError(t, err)

	desc := "some content"
	now := time.Now()

	_, err = store.UpsertItem(ctx, domain.Item{
		FeedID:      feed1.ID,
		GUID:        "1",
		Title:       "Post from X",
		URL:         "https://ex.com/1",
		Description: &desc,
		PublishedAt: &now,
	})
	require.NoError(t, err)

	_, err = store.UpsertItem(ctx, domain.Item{
		FeedID:      feed2.ID,
		GUID:        "2",
		Title:       "Post from Y",
		URL:         "https://ex.com/2",
		Description: &desc,
		PublishedAt: &now,
	})
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items/search?q=Post&feed_ids=1&per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Post from X", resp.Items[0].Title)
}

func TestSearchItems_CaseInsensitive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	store, _, cleanup := pgstoretest.SetupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user, err := store.CreateUser(ctx, "alice", "hash", true)
	require.NoError(t, err)

	feed, err := store.CreateFeed(ctx, user.ID, "https://x.com/rss", "X", "", "", "", "")
	require.NoError(t, err)

	desc := "Searchable CONTENT hErE"
	now := time.Now()

	_, err = store.UpsertItem(ctx, domain.Item{
		FeedID:      feed.ID,
		GUID:        "1",
		Title:       "Hello World",
		URL:         "https://ex.com/1",
		Description: &desc,
		PublishedAt: &now,
	})
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items/search?q=searchable+content&per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
	}

	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Items, 1)
}

func TestItems_Health(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/health", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
