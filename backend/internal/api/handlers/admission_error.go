package handlers

import (
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/models"
	"go.uber.org/zap"
)

func respondAdmissionError(w http.ResponseWriter, logger *zap.Logger, err error) bool {
	status := 0
	switch {
	case errors.Is(err, models.ErrRunQuotaExceeded):
		status = http.StatusTooManyRequests
	case errors.Is(err, models.ErrQueueFull):
		status = http.StatusServiceUnavailable
	default:
		return false
	}
	w.Header().Set("Retry-After", "60")
	respondError(w, logger, status, err.Error())
	return true
}
