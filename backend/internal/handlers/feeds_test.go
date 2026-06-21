package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func TestListFeeds(t *testing.T) {
	store := mockstore.New()
	store.Feeds = []domain.Feed{
		{ID: 1, URL: "https://a.com/rss", Title: "A"},
		{ID: 2, URL: "https://b.com/rss", Title: "B"},
	}

	h := handlers.New(store, nil, nil)
	req := httptest.NewRequest("GET", "/api/feeds", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var feeds []domain.Feed
	err := json.Unmarshal(w.Body.Bytes(), &feeds)
	require.NoError(t, err)
	assert.Len(t, feeds, 2)
}

func TestListFeeds_Empty(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var feeds []domain.Feed
	json.Unmarshal(w.Body.Bytes(), &feeds)
	assert.Empty(t, feeds)
}

func TestAddFeed(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	body := `{"url":"https://example.com/rss"}`
	req := httptest.NewRequest("POST", "/api/feeds", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var feed domain.Feed
	err := json.Unmarshal(w.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/rss", feed.URL)
}

func TestAddFeed_EmptyURL(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	body := `{"url":""}`
	req := httptest.NewRequest("POST", "/api/feeds", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveFeed(t *testing.T) {
	store := mockstore.New()
	store.Feeds = []domain.Feed{{ID: 1, URL: "https://a.com/rss", Title: "A"}}
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("DELETE", "/api/feeds/1", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRemoveFeed_InvalidID(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("DELETE", "/api/feeds/abc", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkFeedRead(t *testing.T) {
	store := mockstore.New()
	store.Feeds = []domain.Feed{{ID: 1, URL: "https://a.com/rss", Title: "A"}}
	store.Items = []domain.Item{
		{ID: 1, FeedID: 1, GUID: "1", Title: "Post 1", URL: "https://ex.com/1", Read: false},
		{ID: 2, FeedID: 1, GUID: "2", Title: "Post 2", URL: "https://ex.com/2", Read: false},
	}
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("POST", "/api/feeds/1/read-all", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, store.Items[0].Read)
	assert.True(t, store.Items[1].Read)
}

func TestMarkFeedRead_InvalidID(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	req := httptest.NewRequest("POST", "/api/feeds/abc/read-all", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
