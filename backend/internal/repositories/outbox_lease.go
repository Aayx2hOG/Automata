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

const outboxDatabaseTimeout = 5 * time.Second

var ErrLeaseLost = errors.New("workflow execution lease lost")

type claimedRun struct {
	id, token uuid.UUID
	graph     models.WorkflowGraph
	seed      map[string]interface{}
}

func (r *pgScheduleRepository) leaseTTL() time.Duration {
	if r.leaseDuration > 0 {
		return r.leaseDuration
	}
	return 30 * time.Second
}

// ListReadyRuns returns hints only. Workers claim after dequeue, so waiting in
// memory never consumes a lease and losing the queue never loses durable work.
func (r *pgScheduleRepository) ListReadyRuns(ctx context.Context, limit int) ([]uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT run_id FROM schedule_outbox
 WHERE lease_until IS NULL OR lease_until <= clock_timestamp()
 ORDER BY created_at, run_id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *pgScheduleRepository) claimRun(ctx context.Context, id uuid.UUID) (*claimedRun, error) {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	claim := &claimedRun{token: uuid.New()}
	var graphJSON, seedJSON []byte
	err = tx.QueryRow(ctx, `SELECT o.run_id, v.graph, o.seed_data FROM schedule_outbox o
 JOIN workflow_runs r ON r.id=o.run_id
 JOIN workflow_versions v ON v.id=r.workflow_version_id
 WHERE ($1::uuid='00000000-0000-0000-0000-000000000000' OR o.run_id=$1)
 AND (o.lease_until IS NULL OR o.lease_until <= clock_timestamp())
 ORDER BY o.created_at, o.run_id LIMIT 1 FOR UPDATE OF o SKIP LOCKED`, id).
		Scan(&claim.id, &graphJSON, &seedJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(graphJSON, &claim.graph); err != nil {
		return nil, err
	}
	if len(seedJSON) > 0 {
		if err = json.Unmarshal(seedJSON, &claim.seed); err != nil {
			return nil, err
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE schedule_outbox SET lease_token=$2,
 lease_until=clock_timestamp()+$3::double precision * interval '1 second' WHERE run_id=$1`,
		claim.id, claim.token, r.leaseTTL().Seconds()); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE workflow_runs SET status='running', started_at=clock_timestamp(),
 finished_at=NULL, error_message=NULL, outputs=NULL WHERE id=$1`, claim.id); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return claim, nil
}

func (r *pgScheduleRepository) renewRun(ctx context.Context, claim *claimedRun) error {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	tag, err := r.pool.Exec(ctx, `UPDATE schedule_outbox
 SET lease_until=clock_timestamp()+$3::double precision * interval '1 second'
 WHERE run_id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()`, claim.id, claim.token, r.leaseTTL().Seconds())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}
	return nil
}

// releaseRun is fenced too: a late shutdown must not reset a replacement worker.
func (r *pgScheduleRepository) releaseRun(claim *claimedRun) {
	ctx, cancel := context.WithTimeout(context.Background(), outboxDatabaseTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return
	} // Expiry recovers work if the database is unavailable.
	defer tx.Rollback(context.Background())
	tag, err := tx.Exec(ctx, `UPDATE schedule_outbox SET lease_token=NULL, lease_until=NULL
 WHERE run_id=$1 AND lease_token=$2`, claim.id, claim.token)
	if err != nil || tag.RowsAffected() == 0 {
		return
	}
	if _, err = tx.Exec(ctx, `UPDATE workflow_runs SET status='pending', started_at=NULL WHERE id=$1`, claim.id); err != nil {
		return
	}
	_ = tx.Commit(ctx)
}

func (r *pgScheduleRepository) finishRun(ctx context.Context, claim *claimedRun, outputs map[string]interface{}, executionErr error) error {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	outputJSON, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	status := models.RunStatusSucceeded
	var message *string
	if executionErr != nil {
		status = models.RunStatusFailed
		msg := executionErr.Error()
		message = &msg
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	// Deletion locks and fences the acknowledgement; rollback preserves the job
	// if storing its terminal result fails. An expired owner cannot acknowledge.
	tag, err := tx.Exec(ctx, `DELETE FROM schedule_outbox
 WHERE run_id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()`, claim.id, claim.token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}
	if _, err = tx.Exec(ctx, `UPDATE workflow_runs SET status=$2, outputs=$3::jsonb,
 error_message=$4, finished_at=clock_timestamp() WHERE id=$1`, claim.id, status, outputJSON, message); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *pgScheduleRepository) ProcessNextRun(ctx context.Context, execute ScheduledExecutor) (bool, error) {
	return r.ProcessRun(ctx, uuid.Nil, execute)
}

// ProcessRun uses short claim/heartbeat/acknowledgement operations. No database
// connection or transaction is held while the executor runs between heartbeats.
func (r *pgScheduleRepository) ProcessRun(ctx context.Context, id uuid.UUID, execute ScheduledExecutor) (bool, error) {
	claim, err := r.claimRun(ctx, id)
	if err != nil || claim == nil {
		return false, err
	}
	finished := false
	defer func() {
		if !finished {
			r.releaseRun(claim)
		}
	}()
	execCtx, cancelExecution := context.WithCancel(ctx)
	defer cancelExecution()
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	done := make(chan struct{})
	var heartbeatErr error
	go func() {
		defer close(done)
		ticker := time.NewTicker(r.leaseTTL() / 3)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				if err := r.renewRun(heartbeatCtx, claim); err != nil {
					if heartbeatCtx.Err() == nil {
						heartbeatErr = fmt.Errorf("renew execution lease: %w", err)
						cancelExecution()
					}
					return
				}
			}
		}
	}()
	defer func() { stopHeartbeat(); <-done }()
	outputs, executionErr := execute(execCtx, claim.graph, claim.seed)
	stopHeartbeat()
	<-done
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if heartbeatErr != nil {
		return false, heartbeatErr
	}
	if err = r.finishRun(ctx, claim, outputs, executionErr); err != nil {
		return false, err
	}
	finished = true
	return true, nil
}
