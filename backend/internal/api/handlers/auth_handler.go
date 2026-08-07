package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/services"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService *services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	req := registerRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || len(req.Password) < 8 {
		respondError(w, h.logger, http.StatusBadRequest, "email and password of atleast 8 characters are required")
		return
	}

	result, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrorUserAlreadyExists) {
			respondError(w, h.logger, http.StatusConflict, "a user with this email already exists")
			return
		}
		h.logger.Error("register failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to register user")
		return
	}
	respondJSON(w, h.logger, http.StatusCreated, AuthResponse{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	req := loginRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrorInvalidCredentials) {
			respondError(w, h.logger, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.logger.Error("login failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to log in")
		return
	}

	respondJSON(w, h.logger, http.StatusOK, AuthResponse{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	req := refreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, models.ErrorInvalidToken) || errors.Is(err, models.ErrorTokenRevoked) {
			respondError(w, h.logger, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		h.logger.Error("refresh failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to refresh refresh token")
		return
	}

	respondJSON(w, h.logger, http.StatusOK, AuthResponse{
		User:         result.User,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	req := refreshRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		h.logger.Error("logout failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to logout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
