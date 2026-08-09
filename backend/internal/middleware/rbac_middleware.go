package middleware

import (
	"net/http"
	"slices"

	"github.com/Aayx2hOG/automata/internal/models"
)

func RequireRole(allowed ...models.Role) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				writeUnAuthorized(w, "missing authentication context")
				return
			}

			if slices.Contains(allowed, role) {
				h.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "insufficient permissions"}`))
		})
	}
}
