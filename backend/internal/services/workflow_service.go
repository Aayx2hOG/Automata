package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultMaxRetries       = 2
	defaultExecutionTimeout = 60 * time.Second
)

type WorkflowService struct {
	workflows repositories.WorkflowRepository
	versions  repositories.WorkflowVersionRepository
	runs      repositories.WorkflowRunRepository
	engine    *workflow.Engine
	logger    *zap.Logger
}

func NewWorkflowService(
	workflows repositories.WorkflowRepository,
	versions repositories.WorkflowVersionRepository,
	runs repositories.WorkflowRunRepository,
	engine *workflow.Engine,
	logger *zap.Logger,
) *WorkflowService {
	return &WorkflowService{
		workflows: workflows,
		versions:  versions,
		runs:      runs,
		engine:    engine,
		logger:    logger,
	}
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.Workflow, *models.WorkflowVersion, error) {
	validation := workflow.ValidatorGraph(graph)
	if !validation.Valid {
		return nil, nil, &workflow.GraphValidationError{Errors: validation.Errors}
	}

	registry := workflow.NewRegistry(s.logger)
	for _, n := range graph.Nodes {
		if _, err := registry.Build(n.Type); err != nil {
			return nil, nil, &workflow.GraphValidationError{Errors: []workflow.ValidationError{{NodeID: n.ID, Message: err.Error()}}}
		}
	}
	wf := &models.Workflow{OwnerID: ownerID, Name: name, Description: description, IsActive: true}
	version, err := s.workflows.CreateWithInitialVersion(ctx, wf, graph)
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

	return s.runs.Enqueue(ctx, wf.ID, version.ID, models.TriggerManual, nil)

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

	return s.runs.Enqueue(ctx, wf.ID, version.ID, models.TriggerWebhook, payload)

}

func flattenOutputs(outputs map[string]map[string]interface{}) map[string]interface{} {
	flattened := make(map[string]interface{}, len(outputs))
	for nodeID, nodeOutput := range outputs {
		flattened[nodeID] = nodeOutput
	}
	return flattened
}

func (s *WorkflowService) CreateSchedule(ctx context.Context, ownerID, workflowID uuid.UUID, cronExpr string) (*models.Schedule, error) {
	if _, err := s.GetWorkflow(ctx, workflowID, ownerID); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *WorkflowService) ExecuteDurableGraph(ctx context.Context, graph models.WorkflowGraph, seed map[string]interface{}) (map[string]interface{}, error) {
	execCtx, cancel := context.WithTimeout(ctx, defaultExecutionTimeout)
	defer cancel()
	for attempt := 0; ; attempt++ {
		var result *workflow.ExecutionResult
		if seed != nil {
			result = s.engine.RunWithSeededOutput(&node.ExecutionContext{Ctx: execCtx}, graph, "webhook_trigger", seed)
		} else {
			result = s.engine.Run(&node.ExecutionContext{Ctx: execCtx}, graph)
		}
		if execCtx.Err() != nil {
			return flattenOutputs(result.Outputs), execCtx.Err()
		}
		if result.Error == nil || attempt == defaultMaxRetries {
			return flattenOutputs(result.Outputs), result.Error
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 2 * time.Second)
		select {
		case <-execCtx.Done():
			timer.Stop()
			return flattenOutputs(result.Outputs), execCtx.Err()
		case <-timer.C:
		}
	}
}

func (s *WorkflowService) VerifyWebhook(ctx context.Context, id uuid.UUID, timestamp, signature string, body []byte) error {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return models.ErrorInvalidToken
	}
	now := time.Now().Unix()
	if seconds < now-300 || seconds > now+300 {
		return models.ErrorInvalidToken
	}
	supplied, err := hex.DecodeString(signature)
	if err != nil {
		return models.ErrorInvalidToken
	}
	wf, err := s.workflows.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrWorkflowNotFound) {
			return models.ErrorInvalidToken
		}
		return err
	}
	if wf.WebhookSecret == "" || !wf.IsActive {
		return models.ErrorInvalidToken
	}
	mac := hmac.New(sha256.New, []byte(wf.WebhookSecret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	if !hmac.Equal(supplied, mac.Sum(nil)) {
		return models.ErrorInvalidToken
	}
	return nil
}

func (s *WorkflowService) GetActiveVersion(ctx context.Context, id, ownerID uuid.UUID) (*models.WorkflowVersion, error) {
	if _, err := s.GetWorkflow(ctx, id, ownerID); err != nil {
		return nil, err
	}
	return s.versions.GetLatestByWorkflow(ctx, id)
}
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.WorkflowVersion, error) {
	if _, err := s.GetWorkflow(ctx, id, ownerID); err != nil {
		return nil, err
	}
	validation := workflow.ValidatorGraph(graph)
	if !validation.Valid {
		return nil, &workflow.GraphValidationError{Errors: validation.Errors}
	}
	registry := workflow.NewRegistry(s.logger)
	for _, n := range graph.Nodes {
		if _, err := registry.Build(n.Type); err != nil {
			return nil, &workflow.GraphValidationError{Errors: []workflow.ValidationError{{NodeID: n.ID, Message: err.Error()}}}
		}
	}
	return s.workflows.UpdateWithVersion(ctx, id, ownerID, name, description, graph)
}
func (s *WorkflowService) ListRuns(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]models.WorkflowRun, error) {
	return s.runs.ListByOwner(ctx, ownerID, limit, offset)
}
