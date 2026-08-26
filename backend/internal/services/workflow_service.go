package services

import (
	"context"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/google/uuid"
)

type WorkflowService struct {
	workflows repositories.WorkflowRepository
	versions  repositories.WorkflowVersionRepository
	runs      repositories.WorkflowRunRepository
	engine    *workflow.Engine
}

func NewWorkflowService(
	workflows repositories.WorkflowRepository,
	versions repositories.WorkflowVersionRepository,
	runs repositories.WorkflowRunRepository,
	engine *workflow.Engine,
) *WorkflowService {
	return &WorkflowService{
		workflows: workflows,
		versions:  versions,
		runs:      runs,
		engine:    engine,
	}
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.Workflow, *models.WorkflowVersion, error) {
	validation := workflow.ValidatorGraph(graph)
	if !validation.Valid {
		return nil, nil, &workflow.GraphValidationError{Errors: validation.Errors}
	}

	wf := &models.Workflow{
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		IsActive:    true,
	}
	if err := s.workflows.Create(ctx, wf); err != nil {
		return nil, nil, err
	}

	version, err := s.versions.Create(ctx, wf.ID, graph)
	if err != nil {
		return nil, nil, err
	}

	return wf, version, nil
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, OwnerID uuid.UUID) ([]models.Workflow, error) {
	return s.workflows.ListByOwner(ctx, OwnerID)
}

func (s *WorkflowService) GetWorkflow(ctx context.Context, id, requesterID uuid.UUID) (*models.Workflow, error) {
	wf, err := s.workflows.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if wf.OwnerID != requesterID {
		return nil, models.ErrWorkflowNotFound
	}
	return wf, nil
}

func (s *WorkflowService) RunWorkflow(ctx context.Context, id, requesterID uuid.UUID) (*models.WorkflowRun, error) {
	wf, err := s.GetWorkflow(ctx, id, requesterID)
	if err != nil {
		return nil, err
	}
	if !wf.IsActive {
		return nil, models.ErrorWorkflowInactive
	}

	version, err := s.versions.GetLatestByWorkflow(ctx, wf.ID)
	if err != nil {
		return nil, err
	}

	run, err := s.runs.Create(ctx, wf.ID, version.ID, models.TriggerManual)
	if err != nil {
		return nil, err
	}

	execCtx := &node.ExecutionContext{Ctx: ctx}
	result := s.engine.Run(execCtx, version.Graph)

	outputs := flattenOutputs(result.Outputs)
	finishedAt := time.Now()

	status := models.RunStatusSucceeded
	var errMsg *string

	if result.Error != nil {
		status = models.RunStatusFailed
		msg := result.Error.Error()
		errMsg = &msg
	}

	if err := s.runs.UpdateResult(ctx, run.ID, status, outputs, errMsg, finishedAt); err != nil {
		return nil, err
	}

	run.Status = status
	run.Outputs = outputs
	run.FinishedAt = &finishedAt
	if errMsg != nil {
		run.ErrorMessage = *errMsg
	}
	return run, nil
}

func flattenOutputs(outputs map[string]map[string]any) map[string]any {
	flattened := make(map[string]any, len(outputs))
	for nodeID, nodeOutput := range outputs {
		flattened[nodeID] = nodeOutput
	}
	return flattened
}

func (s *WorkflowService) GetRun(ctx context.Context, runID, requesterID uuid.UUID) (*models.WorkflowRun, error) {
	run, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetWorkflow(ctx, runID, requesterID); err != nil {
		return nil, err
	}
	return run, err
}
