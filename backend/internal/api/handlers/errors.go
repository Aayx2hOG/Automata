package handlers

import (
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/models"
	"go.uber.org/zap"
)

var errorStatusMap = []struct {
	err    error
	status int
}{
	{models.ErrorUserAlreadyExists, http.StatusConflict},
	{models.ErrorUserNotFound, http.StatusNotFound},
	{models.ErrorInvalidCredentials, http.StatusUnauthorized},
	{models.ErrorInvalidToken, http.StatusUnauthorized},
	{models.ErrorTokenRevoked, http.StatusUnauthorized},
}

func handleAuthErrors(w http.ResponseWriter, logger *zap.Logger, err error) {
	for _, m := range errorStatusMap {
		if errors.Is(err, m.err) {
			logger.Info("auth request rejected", zap.Error(err))
			respondError(w, logger, m.status, m.err.Error())
			return
		}
	}

	logger.Error("unexpected auth error", zap.Error(err))
	respondError(w, logger, http.StatusInternalServerError, "an unexpected error occured")
}
