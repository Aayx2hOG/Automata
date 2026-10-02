package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RunHandler struct {
	workflowService *services.WorkflowService
	logger          *zap.Logger
}

func NewRunHandler(workflowService *services.WorkflowService, logger *zap.Logger) *RunHandler {
	return &RunHandler{workflowService: workflowService, logger: logger}
}

func (h *RunHandler) Get(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify the user")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid run id")
		return
	}

	run, err := h.workflowService.GetRun(r.Context(), id, ownerID)
	if err != nil {
		if errors.Is(err, models.ErrWorkflowNotFound) {
			respondError(w, h.logger, http.StatusNotFound, "run not found")
			return
		}
		h.logger.Error("get run failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to fetch run")
		return
	}
	respondJSON(w, h.logger, http.StatusOK, run)
}

func (h *RunHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}
	limit, offset := 50, 0
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			respondError(w, h.logger, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			respondError(w, h.logger, http.StatusBadRequest, "offset must be non-negative")
			return
		}
	}
	runs, err := h.workflowService.ListRuns(r.Context(), ownerID, limit, offset)
	if err != nil {
		respondError(w, h.logger, http.StatusInternalServerError, "unable to list runs")
		return
	}
	respondJSON(w, h.logger, http.StatusOK, map[string]any{"runs": runs})
}
