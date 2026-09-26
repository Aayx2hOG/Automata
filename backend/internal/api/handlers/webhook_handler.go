package handlers

import (
	"encoding/json"
	"errors"
	"io"
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

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		respondError(w, h.logger, http.StatusRequestEntityTooLarge, "webhook body exceeds limit")
		return
	}
	if err := h.WorkflowService.VerifyWebhook(r.Context(), workflowID, r.Header.Get("X-Webhook-Timestamp"), r.Header.Get("X-Webhook-Signature"), body); err != nil {
		respondError(w, h.logger, http.StatusUnauthorized, "invalid webhook signature")
		return
	}
	payload := map[string]interface{}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid JSON")
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
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
