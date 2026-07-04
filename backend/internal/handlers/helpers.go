package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ekse/rssreader/internal/domain"
)

type feedWithCount struct {
	domain.Feed
	UnreadCount int `json:"unread_count"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
