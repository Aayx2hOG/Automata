package repositories

import (
	"context"
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
