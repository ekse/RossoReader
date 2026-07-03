package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ekse/rssreader/internal/fetcher"
	"github.com/ekse/rssreader/internal/middleware"
	"github.com/ekse/rssreader/internal/scheduler"
	"github.com/ekse/rssreader/internal/store"
)

type Handler struct {
	Store      store.Store
	Scheduler  *scheduler.Scheduler
	Discoverer fetcher.Discoverer
	Auth       *AuthHandler
	Passkey    *PasskeyHandler
}

func New(s store.Store, sch *scheduler.Scheduler, d fetcher.Discoverer, p *PasskeyHandler) *Handler {
	return &Handler{Store: s, Scheduler: sch, Discoverer: d, Auth: NewAuthHandler(s), Passkey: p}
}

// Router returns the public (unauthenticated) routes only. The full router
// with authenticated routes is assembled in cmd/server/main.go so it can wire
// the Authenticate middleware around the protected group.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/api/health", h.Health)

	r.Post("/api/auth/login", h.Auth.Login)
	r.Post("/api/auth/logout", h.Auth.Logout)

	r.Post("/api/auth/passkey/login/begin", h.Passkey.LoginBegin)
	r.Post("/api/auth/passkey/login/finish", h.Passkey.LoginFinish)

	return r
}

// ProtectedRouter returns the routes that require an authenticated user.
func (h *Handler) ProtectedRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/api/auth/me", h.Auth.Me)
	r.Put("/api/auth/password", h.Auth.ChangePassword)

	r.Post("/api/auth/passkey/register/begin", h.Passkey.RegisterBegin)
	r.Post("/api/auth/passkey/register/finish", h.Passkey.RegisterFinish)
	r.Get("/api/auth/passkeys", h.Passkey.ListPasskeys)
	r.Delete("/api/auth/passkeys/{id}", h.Passkey.DeletePasskey)

	r.Get("/api/feeds", h.ListFeeds)
	r.Post("/api/feeds", h.AddFeed)
	r.Post("/api/feeds/discover", h.DiscoverFeeds)
	r.Delete("/api/feeds/{id}", h.RemoveFeed)
	r.Post("/api/feeds/{id}/refresh", h.RefreshFeed)
	r.Post("/api/feeds/{id}/read-all", h.MarkFeedRead)

	r.Get("/api/feeds/opml/export", h.ExportOPML)
	r.Post("/api/feeds/opml/preview", h.PreviewOPMLImport)

	r.Get("/api/items", h.ListItems)
	r.Post("/api/items/read-all", h.MarkAllRead)
	r.Patch("/api/items/{id}", h.UpdateItem)

	r.Get("/api/settings", h.GetSettings)
	r.Patch("/api/settings", h.UpdateSettings)

	return r
}

// AdminRouter returns the routes that require an authenticated admin user.
func (h *Handler) AdminRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/api/users", h.Auth.ListUsers)
	r.Post("/api/users", h.Auth.CreateUser)
	r.Delete("/api/users/{id}", h.Auth.DeleteUser)

	return r
}

// MountRouter assembles the complete chi router with authentication middleware
// around the protected and admin routes. All routes are registered directly on
// one router (chi disallows mounting on an already-mounted path).
func (h *Handler) MountRouter() chi.Router {
	r := chi.NewRouter()

	// Public routes (no auth required).
	r.Get("/api/health", h.Health)
	r.Post("/api/auth/login", h.Auth.Login)
	r.Post("/api/auth/logout", h.Auth.Logout)

	r.Post("/api/auth/passkey/login/begin", h.Passkey.LoginBegin)
	r.Post("/api/auth/passkey/login/finish", h.Passkey.LoginFinish)

	// Authenticated routes.
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(h.Store, h.Auth.cookieName()))

		r.Get("/api/auth/me", h.Auth.Me)
		r.Put("/api/auth/password", h.Auth.ChangePassword)

		r.Post("/api/auth/passkey/register/begin", h.Passkey.RegisterBegin)
		r.Post("/api/auth/passkey/register/finish", h.Passkey.RegisterFinish)
		r.Get("/api/auth/passkeys", h.Passkey.ListPasskeys)
		r.Delete("/api/auth/passkeys/{id}", h.Passkey.DeletePasskey)

		r.Get("/api/feeds", h.ListFeeds)
		r.Post("/api/feeds", h.AddFeed)
		r.Post("/api/feeds/discover", h.DiscoverFeeds)
		r.Delete("/api/feeds/{id}", h.RemoveFeed)
		r.Post("/api/feeds/{id}/refresh", h.RefreshFeed)
		r.Post("/api/feeds/{id}/read-all", h.MarkFeedRead)

		r.Get("/api/feeds/opml/export", h.ExportOPML)
		r.Post("/api/feeds/opml/preview", h.PreviewOPMLImport)

		r.Get("/api/items", h.ListItems)
		r.Post("/api/items/read-all", h.MarkAllRead)
		r.Patch("/api/items/{id}", h.UpdateItem)

		r.Get("/api/settings", h.GetSettings)
		r.Patch("/api/settings", h.UpdateSettings)

		// Admin-only routes.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)
			r.Get("/api/users", h.Auth.ListUsers)
			r.Post("/api/users", h.Auth.CreateUser)
			r.Delete("/api/users/{id}", h.Auth.DeleteUser)
		})
	})

	return r
}

// currentUserID returns the authenticated user ID from the request context.
func currentUserID(r *http.Request) int64 {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		return 0
	}
	return u.ID
}
