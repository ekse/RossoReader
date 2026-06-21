package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func TestGetSettings(t *testing.T) {
	store := mockstore.New()
	store.Settings = map[string]string{"fetch_interval": "30"}

	h := handlers.New(store, nil, nil)
	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settings map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &settings)
	require.NoError(t, err)
	assert.Equal(t, "30", settings["fetch_interval"])
}

func TestUpdateSettings(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil)

	body := `{"fetch_interval":"60"}`
	req := httptest.NewRequest("PATCH", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settings map[string]string
	json.Unmarshal(w.Body.Bytes(), &settings)
	assert.Equal(t, "60", settings["fetch_interval"])
}
