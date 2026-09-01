package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	WorkflowService *services.WorkflowService
	logger          *zap.Logger
}

func NewWebhookHandler(workflowService *services.WorkflowService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{WorkflowService: workflowService, logger: logger}
}

func (h *WebhookHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowID"))
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid workflow id")
		return
	}

	payload := map[string]any{}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}
	if payload == nil {
		payload = map[string]any{}
	}

	run, err := h.WorkflowService.TriggerWebhook(r.Context(), workflowID, payload)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrWorkflowNotFound):
			respondError(w, h.logger, http.StatusNotFound, "not found")
		case errors.Is(err, models.ErrWorkflowVersionNotFound):
			respondError(w, h.logger, http.StatusNotFound, "workflow has no versions")
		case errors.Is(err, models.ErrorWorkflowInactive):
			respondError(w, h.logger, http.StatusConflict, "workflow is inactive")
		default:
			h.logger.Error("webhook trigger failed", zap.Error(err))
			respondError(w, h.logger, http.StatusInternalServerError, "unable to proceed webhook")
		}
		return
	}

	respondJSON(w, h.logger, http.StatusOK, map[string]any{
		"run_id": run.ID,
		"status": run.Status,
	})
}
