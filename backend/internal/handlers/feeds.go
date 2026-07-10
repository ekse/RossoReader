package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/opml"
)

// maxOPMLSize is the maximum allowed OPML file size for import (10 MB).
const maxOPMLSize = 10 << 20

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

	limit, err := h.Store.GetFeedsLimit(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	feeds, err := h.Store.GetFeeds(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(feeds) >= limit {
		http.Error(w, "You have reached the limit of feed subscriptions.", http.StatusConflict)
		return
	}

	feed, err := h.Store.CreateFeed(r.Context(), userID, req.URL, "", "", "", "", "")
	if err != nil {
		if errors.Is(err, domain.ErrFeedAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
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

func (h *Handler) ExportOPML(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	feeds, err := h.Store.GetFeeds(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := opml.GenerateOPML(feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", `attachment; filename="feeds.opml"`)
	w.Write(data)
}

func (h *Handler) PreviewOPMLImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxOPMLSize); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}

	feeds, err := opml.ParseOPML(io.LimitReader(strings.NewReader(string(data)), maxOPMLSize))
	if err != nil {
		http.Error(w, "failed to parse OPML file: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, feeds)
}

func (h *Handler) UnreadCounts(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	counts, err := h.Store.GetUnreadCountByFeed(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"counts": counts})
}
