package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/scheduler"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ScheduleHandler struct {
	workflowService *services.WorkflowService
	schedules       repositories.ScheduleRepository
	scheduler       *scheduler.Scheduler
	logger          *zap.Logger
}

func NewScheduleHandler(workflowService *services.WorkflowService, schedules repositories.ScheduleRepository, sched *scheduler.Scheduler, logger *zap.Logger) *ScheduleHandler {
	return &ScheduleHandler{workflowService: workflowService, schedules: schedules, scheduler: sched, logger: logger}
}

type createScheduleRequest struct {
	CronExpression string `json:"cron_expression" validate:"required"`
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	workflowID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid workflow id")
		return
	}

	req := createScheduleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, validationMessage(err))
		return
	}

	if _, err := h.workflowService.GetWorkflow(r.Context(), workflowID, ownerID); err != nil {
		if errors.Is(err, models.ErrWorkflowNotFound) {
			respondError(w, h.logger, http.StatusNotFound, "workflow not found")
			return
		}
		respondError(w, h.logger, http.StatusInternalServerError, "unable to fetch workflow")
		return
	}

	nextRun, err := h.scheduler.ParseNext(req.CronExpression, time.Now())
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	schedule := &models.Schedule{
		WorkflowID:     workflowID,
		CronExpression: req.CronExpression,
		IsActive:       true,
		NextRunAt:      nextRun,
	}
	if err := h.schedules.Create(r.Context(), schedule); err != nil {
		h.logger.Error("create schedule failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to create schedule")
		return
	}

	respondJSON(w, h.logger, http.StatusCreated, schedule)
}

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	workflowID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid workflow id")
		return
	}

	if _, err := h.workflowService.GetWorkflow(r.Context(), workflowID, ownerID); err != nil {
		respondError(w, h.logger, http.StatusNotFound, "workflow not found")
		return
	}

	schedules, err := h.schedules.ListByWorkflow(r.Context(), workflowID)
	if err != nil {
		h.logger.Error("list schedules failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to list schedules")
		return
	}

	respondJSON(w, h.logger, http.StatusOK, map[string]interface{}{"schedules": schedules})
}

func (h *ScheduleHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, h.logger, http.StatusUnauthorized, "unable to identify user")
		return
	}

	scheduleID, err := uuid.Parse(chi.URLParam(r, "scheduleID"))
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "invalid schedule id")
		return
	}

	if err := h.schedules.Deactivate(r.Context(), scheduleID, ownerID); err != nil {
		if errors.Is(err, models.ErrWorkflowNotFound) {
			respondError(w, h.logger, http.StatusNotFound, "schedule not found")
			return
		}
		h.logger.Error("deactivate schedule failed", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "unable to deactivate schedule")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
