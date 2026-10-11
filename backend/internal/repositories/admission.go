package repositories

import (
	"context"
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AdmissionLimits struct {
	MaxPending        int `mapstructure:"max_pending_runs"`
	MaxPendingPerUser int `mapstructure:"max_pending_per_user"`
	RunsPerMinute     int `mapstructure:"runs_per_minute"`
}

func DefaultAdmissionLimits() AdmissionLimits {
	return AdmissionLimits{MaxPending: 10000, MaxPendingPerUser: 100, RunsPerMinute: 60}
}

// admitRun is called inside the run/outbox transaction for every trigger type.
// A transaction-scoped lock serializes admission across server instances. It is
// held only during enqueue, never while executing a workflow.
func admitRun(ctx context.Context, tx pgx.Tx, workflowID uuid.UUID, limits AdmissionLimits) error {
	if limits == (AdmissionLimits{}) {
		limits = DefaultAdmissionLimits()
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(71829465013001)`); err != nil {
		return err
	}
	var owner uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT owner_id FROM workflows WHERE id=$1`, workflowID).Scan(&owner); err != nil {
		return err
	}
	var total, pending, recent int
	err := tx.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE w.owner_id=$1)
 FROM schedule_outbox o JOIN workflow_runs r ON r.id=o.run_id JOIN workflows w ON w.id=r.workflow_id`, owner).Scan(&total, &pending)
	if err != nil {
		return fmt.Errorf("count pending runs: %w", err)
	}
	if total >= limits.MaxPending {
		return models.ErrQueueFull
	}
	if pending >= limits.MaxPendingPerUser {
		return models.ErrRunQuotaExceeded
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FROM workflow_runs r JOIN workflows w ON w.id=r.workflow_id
 WHERE w.owner_id=$1 AND r.created_at > clock_timestamp()-interval '1 minute'`, owner).Scan(&recent)
	if err != nil {
		return fmt.Errorf("count recent runs: %w", err)
	}
	if recent >= limits.RunsPerMinute {
		return models.ErrRunQuotaExceeded
	}
	return nil
}
