package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/ekse/rossoreader/internal/domain"
	"github.com/ekse/rossoreader/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext returns the authenticated user stored in the context, if any.
func UserFromContext(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(domain.User)
	return u, ok
}

// SetUserInContext stores the user in the context. Useful for tests that want
// to bypass the real cookie-based Authenticate middleware.
func SetUserInContext(ctx context.Context, u domain.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// MustUserFromContext returns the user or the zero value. Use UserFromContext to check presence.
func MustUserFromContext(ctx context.Context) domain.User {
	u, _ := ctx.Value(userContextKey).(domain.User)
	return u
}

// Authenticate reads the session cookie, looks up the session+user via the store,
// and stores the user in the request context. Unauthenticated requests get 401.
func Authenticate(s store.Store, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			uuidID, err := uuid.Parse(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sess, err := s.GetSession(r.Context(), uuidID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, sess.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin requires that the authenticated user is an admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || !u.IsAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
