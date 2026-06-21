package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post 1", URL: "https://ex.com/1", PublishedAt: &now},
		{ID: 2, FeedID: 1, GUID: "2", Title: "Post 2", URL: "https://ex.com/2", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil)
	req := httptest.NewRequest("GET", "/api/items?per_page=10", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

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
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Feed 1 Post", URL: "https://ex.com/1", PublishedAt: &now},
		{ID: 2, FeedID: 2, GUID: "2", Title: "Feed 2 Post", URL: "https://ex.com/2", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil)
	req := httptest.NewRequest("GET", "/api/items?feed_id=1&per_page=10", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []domain.Item `json:"items"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Feed 1 Post", resp.Items[0].Title)
}

func TestUpdateItem_MarkRead(t *testing.T) {
	store := mockstore.New()
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post", URL: "https://ex.com/1", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil)
	body := `{"read":true}`
	req := httptest.NewRequest("PATCH", "/api/items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item domain.Item
	json.Unmarshal(w.Body.Bytes(), &item)
	assert.True(t, item.Read)
}

func TestUpdateItem_MarkStarred(t *testing.T) {
	store := mockstore.New()
	now := time.Now()
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post", URL: "https://ex.com/1", PublishedAt: &now},
	}

	h := handlers.New(store, nil, nil)
	body := `{"starred":true}`
	req := httptest.NewRequest("PATCH", "/api/items/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item domain.Item
	json.Unmarshal(w.Body.Bytes(), &item)
	assert.True(t, item.Starred)
}

func TestHealth(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
