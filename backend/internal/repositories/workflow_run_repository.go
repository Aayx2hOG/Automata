package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkflowRunRepository interface {
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]models.WorkflowRun, error)
	Enqueue(ctx context.Context, workflowID, versionID uuid.UUID, trigger models.TriggerType, seed map[string]interface{}) (*models.WorkflowRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.WorkflowRun, error)
}

type pgWorkflowRunRepository struct {
	pool   *pgxpool.Pool
	limits AdmissionLimits
}

func NewWorkflowRunRepository(pool *pgxpool.Pool, limits ...AdmissionLimits) WorkflowRunRepository {
	r := &pgWorkflowRunRepository{pool: pool, limits: DefaultAdmissionLimits()}
	if len(limits) > 0 {
		r.limits = limits[0]
	}
	return r
}

func (r *pgWorkflowRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, workflow_version_id, status, trigger_type,
		       outputs, error_message, started_at, finished_at, created_at
		FROM workflow_runs
		WHERE id = $1
	`
	var run models.WorkflowRun
	var outputsBytes []byte
	var errMsg *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.WorkflowID, &run.WorkflowVersionID, &run.Status, &run.TriggerType,
		&outputsBytes, &errMsg, &run.StartedAt, &run.FinishedAt, &run.CreatedAt,
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

func (r *pgWorkflowRunRepository) Enqueue(ctx context.Context, workflowID, versionID uuid.UUID, trigger models.TriggerType, seed map[string]interface{}) (*models.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	payload, err := json.Marshal(seed)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	if err := admitRun(ctx, tx, workflowID, r.limits); err != nil {
		return nil, err
	}
	run := &models.WorkflowRun{WorkflowID: workflowID, WorkflowVersionID: versionID, TriggerType: trigger}
	err = tx.QueryRow(ctx, `INSERT INTO workflow_runs(workflow_id,workflow_version_id,trigger_type) VALUES($1,$2,$3) RETURNING id,status,created_at`, workflowID, versionID, trigger).Scan(&run.ID, &run.Status, &run.CreatedAt)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO schedule_outbox(run_id,seed_data) VALUES($1,$2::jsonb)`, run.ID, payload); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return run, nil
}

func (r *pgWorkflowRunRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]models.WorkflowRun, error) {
	rows, err := r.pool.Query(ctx, `SELECT r.id,r.workflow_id,r.workflow_version_id,r.status,r.trigger_type,r.outputs,r.error_message,r.started_at,r.finished_at,r.created_at
 FROM workflow_runs r JOIN workflows w ON w.id=r.workflow_id WHERE w.owner_id=$1 ORDER BY r.created_at DESC,r.id DESC LIMIT $2 OFFSET $3`, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	runs := []models.WorkflowRun{}
	for rows.Next() {
		var run models.WorkflowRun
		var data []byte
		var message *string
		if err := rows.Scan(&run.ID, &run.WorkflowID, &run.WorkflowVersionID, &run.Status, &run.TriggerType, &data, &message, &run.StartedAt, &run.FinishedAt, &run.CreatedAt); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &run.Outputs); err != nil {
				return nil, err
			}
		}
		if message != nil {
			run.ErrorMessage = *message
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}
