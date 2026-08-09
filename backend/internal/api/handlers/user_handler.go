package handlers

import (
	"net/http"

	"github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"go.uber.org/zap"
)

type UserHandler struct {
	users  repositories.UserRepository
	logger *zap.Logger
}

func NewUserHandler(users repositories.UserRepository, logger *zap.Logger) *UserHandler {
	return &UserHandler{users: users, logger: logger}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("fetching current user failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to fetch user")
		return
	}

	respondJSON(w, h.logger, http.StatusOK, user)
}
