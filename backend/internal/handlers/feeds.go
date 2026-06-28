package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/domain"
)

type addFeedRequest struct {
	URL string `json:"url"`
}

func (h *Handler) ListFeeds(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	feeds, err := h.Store.GetFeeds(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	unreadCounts, err := h.Store.GetUnreadCountByFeed(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type feedWithCount struct {
		domain.Feed
		UnreadCount int `json:"unread_count"`
	}

	result := make([]feedWithCount, 0, len(feeds))
	for _, f := range feeds {
		result = append(result, feedWithCount{
			Feed:        f,
			UnreadCount: unreadCounts[f.ID],
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) AddFeed(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	var req addFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	feed, err := h.Store.CreateFeed(r.Context(), userID, req.URL, "", "", "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Scheduler != nil {
		go h.Scheduler.FetchFeedByID(context.Background(), feed.ID)
	}

	writeJSON(w, http.StatusCreated, feed)
}

func (h *Handler) DiscoverFeeds(w http.ResponseWriter, r *http.Request) {
	var req addFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	if h.Discoverer == nil {
		http.Error(w, "discoverer not available", http.StatusServiceUnavailable)
		return
	}

	feeds, err := h.Discoverer.Discover(r.Context(), req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, feeds)
}

func (h *Handler) RemoveFeed(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteFeed(r.Context(), userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	if h.Scheduler == nil {
		http.Error(w, "scheduler not available", http.StatusServiceUnavailable)
		return
	}

	// Validate ownership before fetching.
	if _, err := h.Store.GetFeed(r.Context(), userID, id); err != nil {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}

	if err := h.Scheduler.FetchFeedByID(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) MarkFeedRead(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	if err := h.Store.MarkAllFeedItemsRead(r.Context(), userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
