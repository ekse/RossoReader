package handlers_test

import (
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

	h := handlers.New(store, nil, nil)
	r := authedRouter(h, user)

	req := authReq("GET", "/api/items?per_page=10", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
		Total int64          `json:"total"`
		Page  int            `json:"page"`
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

	h := handlers.New(store, nil, nil)
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

	h := handlers.New(store, nil, nil)
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

	h := handlers.New(store, nil, nil)
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

func TestItems_Health(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", true)
	h := handlers.New(store, nil, nil)
	r := authedRouter(h, user)

	req := authReq("GET", "/api/health", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}