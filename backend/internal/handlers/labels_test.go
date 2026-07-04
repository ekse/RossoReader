package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func TestListLabels_Empty(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/labels", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var labels []domain.Label
	json.Unmarshal(w.Body.Bytes(), &labels)
	assert.Empty(t, labels)
}

func TestCreateLabel(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/labels", `{"name":"Tech"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var label domain.Label
	err := json.Unmarshal(w.Body.Bytes(), &label)
	require.NoError(t, err)
	assert.Equal(t, "Tech", label.Name)
	assert.Equal(t, user.ID, label.UserID)
}

func TestCreateLabel_Duplicate(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	store.NextLabelID = 2
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/labels", `{"name":"Tech"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateLabel_EmptyName(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/labels", `{"name":""}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLabels(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Labels = []domain.Label{
		{ID: 1, UserID: user.ID, Name: "News"},
		{ID: 2, UserID: user.ID, Name: "Tech"},
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/labels", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var labels []domain.Label
	err := json.Unmarshal(w.Body.Bytes(), &labels)
	require.NoError(t, err)
	assert.Len(t, labels, 2)
}

func TestDeleteLabel(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("DELETE", "/api/labels/1", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	labels, _ := store.GetLabels(context.Background(), user.ID)
	assert.Empty(t, labels)
}

func TestUpdateLabel(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("PUT", "/api/labels/1", `{"name":"Technology"}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	label, err := store.GetLabel(context.Background(), user.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "Technology", label.Name)
}

func TestAddFeedLabel(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("POST", "/api/feeds/1/labels", `{"label_id":1}`, user)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	labels, _ := store.GetFeedLabels(context.Background(), user.ID, 1)
	assert.Len(t, labels, 1)
	assert.Equal(t, "Tech", labels[0].Name)
}

func TestRemoveFeedLabel(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	store.FeedLabels = map[int64][]int64{1: {1}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("DELETE", "/api/feeds/1/labels/1", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	labels, _ := store.GetFeedLabels(context.Background(), user.ID, 1)
	assert.Empty(t, labels)
}

func TestGetFeedLabels(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Feeds = []domain.Feed{{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"}}
	store.Labels = []domain.Label{{ID: 1, UserID: user.ID, Name: "Tech"}}
	store.FeedLabels = map[int64][]int64{1: {1}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/1/labels", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var labels []domain.Label
	err := json.Unmarshal(w.Body.Bytes(), &labels)
	require.NoError(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, "Tech", labels[0].Name)
}

func TestAddFeedLabel_WrongOwnership(t *testing.T) {
	store := mockstore.New()
	userA := makeUser(t, store, "alice", false)
	userB := makeUser(t, store, "bob", false)
	store.Feeds = []domain.Feed{{ID: 1, UserID: userA.ID, URL: "https://a.com/rss", Title: "A"}}
	store.Labels = []domain.Label{{ID: 1, UserID: userA.ID, Name: "Tech"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, userB)

	// User B cannot assign label from user A to a feed.
	req := authReq("POST", "/api/feeds/1/labels", `{"label_id":1}`, userB)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLabelOwnership(t *testing.T) {
	store := mockstore.New()
	userA := makeUser(t, store, "alice", false)
	userB := makeUser(t, store, "bob", false)
	store.Labels = []domain.Label{
		{ID: 1, UserID: userA.ID, Name: "AliceLabel"},
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, userB)

	// User B should not see User A's labels.
	req := authReq("GET", "/api/labels", "", userB)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var labels []domain.Label
	json.Unmarshal(w.Body.Bytes(), &labels)
	assert.Empty(t, labels)
}

func TestDeleteLabel_WrongUser(t *testing.T) {
	store := mockstore.New()
	userA := makeUser(t, store, "alice", false)
	userB := makeUser(t, store, "bob", false)
	store.Labels = []domain.Label{{ID: 1, UserID: userA.ID, Name: "AliceLabel"}}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, userB)

	req := authReq("DELETE", "/api/labels/1", "", userB)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// No error since the WHERE clause matches nothing.
	assert.Equal(t, http.StatusNoContent, w.Code)
	// Label should still exist for user A.
	labels, _ := store.GetLabels(context.Background(), userA.ID)
	assert.Len(t, labels, 1)
}

func TestCreateLabel_SameNameDifferentUsers(t *testing.T) {
	store := mockstore.New()
	userA := makeUser(t, store, "alice", false)
	userB := makeUser(t, store, "bob", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))

	// User A creates "Tech".
	rA := authedRouter(h, userA)
	reqA := authReq("POST", "/api/labels", `{"name":"Tech"}`, userA)
	reqA.Header.Set("Content-Type", "application/json")
	wA := httptest.NewRecorder()
	rA.ServeHTTP(wA, reqA)
	assert.Equal(t, http.StatusCreated, wA.Code)

	// User B creates "Tech" — same name, different user, should succeed.
	rB := authedRouter(h, userB)
	reqB := authReq("POST", "/api/labels", `{"name":"Tech"}`, userB)
	reqB.Header.Set("Content-Type", "application/json")
	wB := httptest.NewRecorder()
	rB.ServeHTTP(wB, reqB)
	assert.Equal(t, http.StatusCreated, wB.Code)
}

func TestGroupedFeeds_NoLabels(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"},
		{ID: 2, UserID: user.ID, URL: "https://b.com/rss", Title: "B"},
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/grouped", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	groups := resp["label_groups"].([]any)
	assert.Empty(t, groups)

	unlabeled := resp["unlabeled_feeds"].([]any)
	assert.Len(t, unlabeled, 2)
}

func TestGroupedFeeds_WithLabels(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	store.Feeds = []domain.Feed{
		{ID: 1, UserID: user.ID, URL: "https://a.com/rss", Title: "A"},
		{ID: 2, UserID: user.ID, URL: "https://b.com/rss", Title: "B"},
		{ID: 3, UserID: user.ID, URL: "https://c.com/rss", Title: "C"},
	}
	store.Labels = []domain.Label{
		{ID: 1, UserID: user.ID, Name: "Tech"},
		{ID: 2, UserID: user.ID, Name: "News"},
	}
	store.FeedLabels = map[int64][]int64{
		1: {1}, // feed 1 has label 1 (Tech)
		2: {1}, // feed 2 has label 1 (Tech)
		3: {2}, // feed 3 has label 2 (News)
	}
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := authReq("GET", "/api/feeds/grouped", "", user)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	groups := resp["label_groups"].([]any)
	require.Len(t, groups, 2)

	unlabeled := resp["unlabeled_feeds"].([]any)
	assert.Empty(t, unlabeled)
}
