package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func HealthCheck(pool *pgxpool.Pool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, logger, map[string]string{"status": "error", "db": "unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		writeJSON(w, logger, map[string]string{
			"Status": "ok",
			"db":     "connected",
		})
	}
}

func writeJSON(w http.ResponseWriter, logger *zap.Logger, payload map[string]string) {
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("failed to encode response:", zap.Error(err))
	}
}
