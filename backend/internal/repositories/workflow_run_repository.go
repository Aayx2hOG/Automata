package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkflowRunRepository interface {
	Create(ctx context.Context, workflowID, workflowVersionID uuid.UUID, triggerType models.TriggerType) (*models.WorkflowRun, error)
	UpdateResult(ctx context.Context, runID uuid.UUID, status models.RunStatus, outputs map[string]any, errMsg *string, finishedAt time.Time) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.WorkflowRun, error)
}

type pgWorkflowRunRepository struct {
	pool *pgxpool.Pool
}

func NewWorkflowRunRepository(pool *pgxpool.Pool) WorkflowRunRepository {
	return &pgWorkflowRunRepository{pool: pool}
}

func (r *pgWorkflowRunRepository) Create(ctx context.Context, workflowID, workflowVersionID uuid.UUID, triggerType models.TriggerType) (*models.WorkflowRun, error) {
	query := `
		INSERT INTO workflow_runs (workflow_id, workflow_version_id, status, trigger_type, started_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, status, started_at, created_at
	`
	run := &models.WorkflowRun{
		WorkflowID:        workflowID,
		WorkflowVersionID: workflowVersionID,
		TriggerType:       triggerType,
	}
	err := r.pool.QueryRow(ctx, query, workflowID, workflowVersionID, models.RunStatusRunning, triggerType).Scan(&run.ID, &run.Status, &run.StartedAt, &run.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert workflow run: %w", err)
	}

	return run, nil
}

func (r *pgWorkflowRunRepository) UpdateResult(ctx context.Context, runID uuid.UUID, status models.RunStatus, outputs map[string]any, errMsg *string, finishedAt time.Time) error {
	outputsJSON := []byte{}
	if outputs != nil {
		marshaled, err := json.Marshal(outputs)
		if err != nil {
			return fmt.Errorf("marshal run outputs: %w", err)
		}
		outputsJSON = marshaled
	}
	query := `
		UPDATE workflow_runs
		SET status = $1, outputs = $2::jsonb, error_message = $3, finished_at = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, status, outputsJSON, errMsg, finishedAt, runID)
	if err != nil {
		return fmt.Errorf("update workflow run result: %w", err)
	}
	return nil
}

func (r *pgWorkflowRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, workflow_version_id, status, trigger_type,
		       outputs, error_message, started_at, finished_at, created_at
		FROM workflow_runs
		WHERE id = $1
	`
	run := models.WorkflowRun{}
	outputsBytes := []byte{}
	var errMsg *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.WorkflowVersionID, &run.Status, &run.TriggerType, &outputsBytes, &errMsg, &run.StartedAt, &run.FinishedAt, &run.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrWorkflowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query workflow run: %w", err)
	}

	if outputsBytes != nil {
		if err := json.Unmarshal(outputsBytes, &run.Outputs); err != nil {
			return nil, fmt.Errorf("unmarshal run outputs: %w", err)
		}
	}
	if errMsg != nil {
		run.ErrorMessage = *errMsg
	}

	return &run, nil
}
