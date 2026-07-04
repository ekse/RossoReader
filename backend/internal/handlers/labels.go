package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/domain"
)

type createLabelRequest struct {
	Name string `json:"name"`
}

type updateLabelRequest struct {
	Name string `json:"name"`
}

type addFeedLabelRequest struct {
	LabelID int64 `json:"label_id"`
}

type labelGroup struct {
	Label domain.Label    `json:"label"`
	Feeds []feedWithCount `json:"feeds"`
}

func (h *Handler) ListLabels(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	labels, err := h.Store.GetLabels(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, labels)
}

func (h *Handler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	var req createLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	label, err := h.Store.CreateLabel(r.Context(), userID, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrLabelAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, label)
}

func (h *Handler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid label id", http.StatusBadRequest)
		return
	}

	var req updateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := h.Store.UpdateLabel(r.Context(), userID, id, req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid label id", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteLabel(r.Context(), userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetFeedLabels(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	feedIDStr := chi.URLParam(r, "id")
	feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	labels, err := h.Store.GetFeedLabels(r.Context(), userID, feedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, labels)
}

func (h *Handler) AddFeedLabel(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	feedIDStr := chi.URLParam(r, "id")
	feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	var req addFeedLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Store.AddFeedLabel(r.Context(), userID, feedID, req.LabelID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveFeedLabel(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	feedIDStr := chi.URLParam(r, "id")
	feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}
	labelIDStr := chi.URLParam(r, "lid")
	labelID, err := strconv.ParseInt(labelIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid label id", http.StatusBadRequest)
		return
	}

	if err := h.Store.RemoveFeedLabel(r.Context(), userID, feedID, labelID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GroupedFeeds(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)

	feeds, err := h.Store.GetFeeds(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	labels, err := h.Store.GetLabels(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feedIDsByLabel, err := h.Store.GetFeedIDsByLabel(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	unreadCounts, err := h.Store.GetUnreadCountByFeed(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feedByID := make(map[int64]feedWithCount, len(feeds))
	labeledFeedIDs := make(map[int64]bool)
	for _, f := range feeds {
		feedByID[f.ID] = feedWithCount{Feed: f, UnreadCount: unreadCounts[f.ID]}
	}

	labelGroups := make([]labelGroup, 0, len(labels))
	for _, l := range labels {
		feedIDs := feedIDsByLabel[l.ID]
		groupFeeds := make([]feedWithCount, 0, len(feedIDs))
		for _, fid := range feedIDs {
			if f, ok := feedByID[fid]; ok {
				groupFeeds = append(groupFeeds, f)
				labeledFeedIDs[fid] = true
			}
		}
		labelGroups = append(labelGroups, labelGroup{Label: l, Feeds: groupFeeds})
	}

	unlabeled := make([]feedWithCount, 0, len(feeds))
	for _, f := range feeds {
		if !labeledFeedIDs[f.ID] {
			unlabeled = append(unlabeled, feedByID[f.ID])
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"label_groups":    labelGroups,
		"unlabeled_feeds": unlabeled,
	})
}
