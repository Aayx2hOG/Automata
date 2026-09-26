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
)

// ScheduledExecutor returns a workflow outcome. Context cancellation instead
// leaves durable work pending, so shutdown does not discard an occurrence.
type ScheduledExecutor func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error)

// EnqueueRun atomically claims an occurrence, creates its run and durable work,
// and advances the schedule. No ownership state survives a rolled-back transaction.
func (r *pgScheduleRepository) EnqueueRun(ctx context.Context, schedule models.Schedule, ranAt, nextRunAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	var workflowID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT workflow_id FROM schedules
  WHERE id = $1 AND is_active AND next_run_at = $2 AND next_run_at <= $3
  FOR UPDATE SKIP LOCKED`, schedule.ID, schedule.NextRunAt, ranAt).Scan(&workflowID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("claim schedule: %w", err)
	}
	if !nextRunAt.After(ranAt) {
		return fmt.Errorf("next run must be after current time")
	}

	var versionID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT v.id FROM workflow_versions v
  JOIN workflows w ON w.id = v.workflow_id
  WHERE w.id = $1 AND w.is_active ORDER BY v.version DESC LIMIT 1`, workflowID).Scan(&versionID)
	if err != nil {
		return fmt.Errorf("scheduled workflow version: %w", err)
	}
	var runID uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO workflow_runs
  (workflow_id, workflow_version_id, trigger_type, schedule_id, scheduled_for)
  VALUES ($1, $2, 'cron', $3, $4) RETURNING id`, workflowID, versionID, schedule.ID, schedule.NextRunAt).Scan(&runID)
	if err != nil {
		return fmt.Errorf("create scheduled run: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO schedule_outbox (run_id) VALUES ($1)`, runID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE schedules SET last_run_at = $2, next_run_at = $3 WHERE id = $1`, schedule.ID, ranAt, nextRunAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ProcessNextRun holds a database lock until execution and acknowledgement finish.
// Other schedulers skip that row; connection loss releases it for retry. This uses
// one connection per worker, with execution bounded by the service timeout.
// External node side effects remain at-least-once across process crashes.
func (r *pgScheduleRepository) ProcessNextRun(ctx context.Context, execute ScheduledExecutor) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.Background())
	var runID uuid.UUID
	var graphJSON, seedJSON []byte
	err = tx.QueryRow(ctx, `SELECT o.run_id, v.graph, o.seed_data FROM schedule_outbox o
  JOIN workflow_runs r ON r.id = o.run_id
  JOIN workflow_versions v ON v.id = r.workflow_version_id
  ORDER BY o.created_at, o.run_id LIMIT 1 FOR UPDATE OF o SKIP LOCKED`).Scan(&runID, &graphJSON, &seedJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var graph models.WorkflowGraph
	if err = json.Unmarshal(graphJSON, &graph); err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE workflow_runs SET status = 'running', started_at = clock_timestamp() WHERE id = $1`, runID); err != nil {
		return false, err
	}
	var seed map[string]interface{}
	if len(seedJSON) > 0 {
		if err = json.Unmarshal(seedJSON, &seed); err != nil {
			return false, err
		}
	}
	outputs, executionErr := execute(ctx, graph, seed)
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	status := models.RunStatusSucceeded
	var message *string
	if executionErr != nil {
		status = models.RunStatusFailed
		msg := executionErr.Error()
		message = &msg
	}
	outputJSON, err := json.Marshal(outputs)
	if err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE workflow_runs SET status = $2, outputs = $3::jsonb,
  error_message = $4, finished_at = clock_timestamp() WHERE id = $1`, runID, status, outputJSON, message); err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM schedule_outbox WHERE run_id = $1`, runID); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
