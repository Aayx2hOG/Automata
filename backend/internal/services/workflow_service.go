package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/queue"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkflowService struct {
	workflows repositories.WorkflowRepository
	versions  repositories.WorkflowVersionRepository
	runs      repositories.WorkflowRunRepository
	engine    *workflow.Engine
	queue     *queue.Queue
	logger    *zap.Logger
}

func NewWorkflowService(
	workflows repositories.WorkflowRepository,
	versions repositories.WorkflowVersionRepository,
	runs repositories.WorkflowRunRepository,
	engine *workflow.Engine,
	q *queue.Queue,
	logger *zap.Logger,
) *WorkflowService {
	return &WorkflowService{
		workflows: workflows,
		versions:  versions,
		runs:      runs,
		engine:    engine,
		queue:     q,
		logger:    logger,
	}
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.Workflow, *models.WorkflowVersion, error) {
	validation := workflow.ValidatorGraph(graph)
	if !validation.Valid {
		return nil, nil, &workflow.GraphValidationError{Errors: validation.Errors}
	}

	wf := &models.Workflow{OwnerID: ownerID, Name: name, Description: description, IsActive: true}
	if err := s.workflows.Create(ctx, wf); err != nil {
		return nil, nil, err
	}

	version, err := s.versions.Create(ctx, wf.ID, graph)
	if err != nil {
		return nil, nil, err
	}

	return wf, version, nil
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, ownerID uuid.UUID) ([]models.Workflow, error) {
	return s.workflows.ListByOwner(ctx, ownerID)
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

func (s *WorkflowService) GetRun(ctx context.Context, runID, requesterID uuid.UUID) (*models.WorkflowRun, error) {
	run, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetWorkflow(ctx, run.WorkflowID, requesterID); err != nil {
		return nil, err
	}
	return run, nil
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

	if !s.queue.Enqueue(queue.Job{RunID: run.ID, Graph: version.Graph}) {
		return s.failEnqueue(ctx, run)
	}

	return run, nil
}

func (s *WorkflowService) TriggerWebhook(ctx context.Context, workflowID uuid.UUID, payload map[string]interface{}) (*models.WorkflowRun, error) {
	wf, err := s.workflows.GetById(ctx, workflowID)
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

	run, err := s.runs.Create(ctx, wf.ID, version.ID, models.TriggerWebhook)
	if err != nil {
		return nil, err
	}

	job := queue.Job{RunID: run.ID, Graph: version.Graph, SeedNodeType: "webhook_trigger", SeedData: payload}
	if !s.queue.Enqueue(job) {
		return s.failEnqueue(ctx, run)
	}

	return run, nil
}

func (s *WorkflowService) failEnqueue(ctx context.Context, run *models.WorkflowRun) (*models.WorkflowRun, error) {
	msg := "job queue is full, try again shortly"
	finishedAt := time.Now()
	if err := s.runs.UpdateResult(ctx, run.ID, models.RunStatusFailed, nil, &msg, finishedAt); err != nil {
		s.logger.Error("failed to record enqueue failure", zap.Error(err))
	}
	return nil, fmt.Errorf("job queue is full")
}

func (s *WorkflowService) ExecuteJob(ctx context.Context, job queue.Job) {
	if err := s.runs.MarkRunning(ctx, job.RunID); err != nil {
		s.logger.Error("failed to mark run running", zap.Error(err), zap.String("run_id", job.RunID.String()))
	}

	execCtx := &node.ExecutionContext{Ctx: ctx}

	var result *workflow.ExecutionResult
	if job.SeedNodeType != "" {
		result = s.engine.RunWithSeededOutput(execCtx, job.Graph, job.SeedNodeType, job.SeedData)
	} else {
		result = s.engine.Run(execCtx, job.Graph)
	}

	outputs := flattenOutputs(result.Outputs)
	finishedAt := time.Now()

	status := models.RunStatusSucceeded
	var errMsg *string
	if result.Error != nil {
		status = models.RunStatusFailed
		msg := result.Error.Error()
		errMsg = &msg
	}

	if err := s.runs.UpdateResult(ctx, job.RunID, status, outputs, errMsg, finishedAt); err != nil {
		s.logger.Error("failed to record run result", zap.Error(err), zap.String("run_id", job.RunID.String()))
	}
}

func flattenOutputs(outputs map[string]map[string]interface{}) map[string]interface{} {
	flattened := make(map[string]interface{}, len(outputs))
	for nodeID, nodeOutput := range outputs {
		flattened[nodeID] = nodeOutput
	}
	return flattened
}
