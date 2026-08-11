package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func HealthCheck(pool *pgxpool.Pool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			respondJSON(w, logger, http.StatusOK, map[string]string{"status": "ok", "db": "not_configured"})
			return
		}

		if err := pool.Ping(r.Context()); err != nil {
			respondJSON(w, logger, http.StatusServiceUnavailable, map[string]string{"status": "error", "db": "unreachable"})
			return
		}
		respondJSON(w, logger, http.StatusOK, map[string]string{
			"status": "ok",
			"db":     "connected",
		})
	}
}
