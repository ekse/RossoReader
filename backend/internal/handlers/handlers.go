package handlers

import (
	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/scheduler"
	"github.com/ekse/rssreader/internal/store"
)

type Handler struct {
	Store     store.Store
	Scheduler *scheduler.Scheduler
}

func New(s store.Store, sch *scheduler.Scheduler) *Handler {
	return &Handler{Store: s, Scheduler: sch}
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/api/feeds", h.ListFeeds)
	r.Post("/api/feeds", h.AddFeed)
	r.Delete("/api/feeds/{id}", h.RemoveFeed)
	r.Post("/api/feeds/{id}/refresh", h.RefreshFeed)
	r.Post("/api/feeds/{id}/read-all", h.MarkFeedRead)

	r.Get("/api/items", h.ListItems)
	r.Post("/api/items/read-all", h.MarkAllRead)
	r.Patch("/api/items/{id}", h.UpdateItem)

	r.Get("/api/settings", h.GetSettings)
	r.Patch("/api/settings", h.UpdateSettings)

	r.Get("/api/health", h.Health)

	return r
}
