package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/domain"
)

type listItemsResponse struct {
	Items []domain.Item `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
}

type updateItemRequest struct {
	Read    *bool `json:"read,omitempty"`
	Starred *bool `json:"starred,omitempty"`
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var feedID *int64
	if fid := q.Get("feed_id"); fid != "" {
		v, err := strconv.ParseInt(fid, 10, 64)
		if err == nil {
			feedID = &v
		}
	}

	var read *bool
	if rv := q.Get("read"); rv != "" {
		v := rv == "true"
		read = &v
	}

	var starred *bool
	if sv := q.Get("starred"); sv != "" {
		v := sv == "true"
		starred = &v
	}

	query := domain.ItemsQuery{
		Page:    page,
		PerPage: perPage,
		UserID:  userID,
		FeedID:  feedID,
		Read:    read,
		Starred: starred,
	}

	items, total, err := h.Store.GetItems(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, listItemsResponse{
		Items: items,
		Total: total,
		Page:  page,
	})
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}

	var req updateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Read != nil {
		if err := h.Store.MarkItemRead(r.Context(), userID, id, *req.Read); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Starred != nil {
		if err := h.Store.MarkItemStarred(r.Context(), userID, id, *req.Starred); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	item, err := h.Store.GetItem(r.Context(), userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	if err := h.Store.MarkAllItemsRead(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
