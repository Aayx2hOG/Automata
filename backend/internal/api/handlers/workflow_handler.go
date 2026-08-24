package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkflowHandler struct {
	workflowService *services.WorkflowService
	logger          *zap.Logger
}

func NewWorkflowHandler(workflowService *services.WorkflowService, logger *zap.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		logger:          logger,
	}
}

type createWorkflowRequest struct {
	Name        string               `json:"name" validate:"required.max=255"`
	Description string               `json:"description" validate:"max=1000"`
	Graph       models.WorkflowGraph `json:"graph"`
}

type workflowValidationErrorResponse struct {
	Error  string                     `json:"error:"`
	Errors []workflow.ValidationError `json:"errors"`
}

func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	req := createWorkflowRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	wf, version, err := h.workflowService.CreateWorkflow(r.Context(), ownerID, req.Name, req.Description, req.Graph)
	if err != nil {
		var graphErr *workflow.GraphValidationError
		if errors.As(err, &graphErr) {
			respondJSON(w, h.logger, http.StatusBadRequest, workflowValidationErrorResponse{
				Error:  "workflow graph is invalid",
				Errors: graphErr.Errors,
			})
			return
		}
		h.logger.Error("create worflow failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to create workflow")
		return
	}
	respondJSON(w, h.logger, http.StatusCreated, map[string]any{
		"Workflow": wf,
		"Version":  version,
	})
}

func (h *WorkflowHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	workflows, err := h.workflowService.ListWorkflows(r.Context(), ownerID)
	if err != nil {
		h.logger.Error("list workflows failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to list workflows")
		return
	}

	respondJSON(w, h.logger, http.StatusOK, map[string]any{
		"workflows": workflows,
	})
}

func (h *WorkflowHandler) Get(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	id, err := parseWorkfloeID(r)
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid workflow id")
		return
	}

	wf, err := h.workflowService.GetWorkflow(r.Context(), id, ownerID)
	if err != nil {
		if errors.Is(err, models.ErrWorkflowNotFound) {
			respondError(w, h.logger, http.StatusNotFound, "workflow not found")
			return
		}

		h.logger.Error("get workflow failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to fetch workflow")
		return
	}
	respondJSON(w, h.logger, http.StatusOK, wf)
}

func (h *WorkflowHandler) Run(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	id, err := parseWorkfloeID(r)
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid workflow id")
		return
	}

	run, err := h.workflowService.RunWorkflow(r.Context(), id, ownerID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrWorkflowNotFound):
			respondError(w, h.logger, http.StatusNotFound, "workflow not found")
		case errors.Is(err, models.ErrorWorkflowInactive):
			respondError(w, h.logger, http.StatusConflict, "workflow is inactive")
		default:
			h.logger.Error("run workflow failed", zap.Error(err))
			respondError(w, h.logger, http.StatusInternalServerError, "unable to run workflow")
		}
		return
	}
	respondJSON(w, h.logger, http.StatusOK, run)
}

func parseWorkfloeID(r *http.Request) (uuid.UUID, error) {
	idParam := chi.URLParam(r, "id")
	return uuid.Parse(idParam)
}
