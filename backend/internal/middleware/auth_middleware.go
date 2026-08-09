package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Aayx2hOG/automata/internal/auth"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDContextKey   contextKey = "userID"
	userRoleContextKey contextKey = "userRole"
)

func RequireAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnAuthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnAuthorized(w, "authorization header must be in the form: Bearer <token>")
				return
			}

			claims, err := jwtManager.VerifyAccessToken(parts[1])
			if err != nil {
				writeUnAuthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, userRoleContextKey, claims.Role)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnAuthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}

func RoleFromContext(ctx context.Context) (models.Role, bool) {
	role, ok := ctx.Value(userRoleContextKey).(models.Role)
	return role, ok
}
