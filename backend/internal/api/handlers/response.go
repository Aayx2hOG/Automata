package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func respondJSON(w http.ResponseWriter, logger *zap.Logger, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
	}
}

func respondError(w http.ResponseWriter, logger *zap.Logger, status int, message string) {
	respondJSON(w, logger, status, map[string]string{"error": message})
}
