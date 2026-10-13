package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/credentials"
	"github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CredentialHandler(vault *credentials.Vault, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		owner, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			respondError(w, logger, 401, "unable to identify user")
			return
		}
		if vault == nil {
			respondError(w, logger, 503, "credential storage is not configured")
			return
		}
		id := uuid.Nil
		if raw := chi.URLParam(r, "id"); raw != "" {
			var err error
			id, err = uuid.Parse(raw)
			if err != nil || id == uuid.Nil {
				respondError(w, logger, 400, "invalid credential id")
				return
			}
		}
		var result any
		var err error
		status := http.StatusOK
		switch r.Method {
		case http.MethodGet:
			result, err = vault.List(r.Context(), owner)
		case http.MethodDelete:
			err = vault.Delete(r.Context(), owner, id)
			status = http.StatusNoContent
		case http.MethodPost, http.MethodPut:
			var input credentials.Input
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				respondError(w, logger, 400, "invalid credential request")
				return
			}
			if err := input.Validate(); err != nil {
				respondError(w, logger, 400, err.Error())
				return
			}
			result, err = vault.Save(r.Context(), owner, id, input)
			if r.Method == http.MethodPost {
				status = http.StatusCreated
			}
		}
		if errors.Is(err, credentials.ErrNotFound) {
			respondError(w, logger, 404, "credential not found")
			return
		}
		if err != nil {
			respondError(w, logger, 500, "unable to access credential storage")
			return
		}
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		respondJSON(w, logger, status, result)
	}
}
