package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate = validator.New()

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
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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

	if err := validate.Struct(req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	result, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		handleAuthErrors(w, h.logger, err)
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

	if err := validate.Struct(req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleAuthErrors(w, h.logger, err)
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

	if err := validate.Struct(req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	result, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		handleAuthErrors(w, h.logger, err)
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

	if err := validate.Struct(req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		handleAuthErrors(w, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validationMessage(err error) string {
	ve := validator.ValidationErrors{}
	if errors.As(err, &ve) && len(ve) > 0 {
		f := ve[0]
		switch f.Tag() {
		case "required":
			return f.Field() + " is required"
		case "email":
			return "email must be a valid email addresss"
		case "min":
			return f.Field() + " must be atleast " + f.Param() + " characters"
		case "max":
			return f.Field() + " must be atleast " + f.Param() + " characters"
		default:
			return f.Field() + " is invalid"
		}
	}
	return "invalid request"
}
