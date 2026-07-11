package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ekse/rossoreader/internal/domain"
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

// parseIntList parses a comma-separated string of integers (e.g. "1,5,12")
// into a slice of int64. Invalid or empty entries are silently skipped.
// Returns nil when the input string is empty.
func parseIntList(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
