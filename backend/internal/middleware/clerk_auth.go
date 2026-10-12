package middleware

import (
	"context"
	"github.com/Aayx2hOG/automata/internal/models"
	"net/http"
	"strings"
)

type SessionAuthenticator interface {
	Authenticate(context.Context, string) (*models.User, error)
}

func RequireClerkAuth(auth SessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnAuthorized(w, "missing or invalid authorization header")
				return
			}
			user, err := auth.Authenticate(r.Context(), parts[1])
			if err != nil || user == nil || !user.IsActive {
				writeUnAuthorized(w, "unable to authenticate account")
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey, user.ID)
			ctx = context.WithValue(ctx, userRoleContextKey, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
