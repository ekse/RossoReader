package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	settings, err := h.Store.GetSettings(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	for key, value := range req {
		if err := h.Store.UpsertSetting(r.Context(), userID, key, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	settings, err := h.Store.GetSettings(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}