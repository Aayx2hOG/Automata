package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/jackc/pgx/v5"
)

type runNodeState struct {
	repository *pgScheduleRepository
	claim      *claimedRun
}

func (s *runNodeState) IdempotencyKey(nodeID string) string {
	sum := sha256.Sum256([]byte(s.claim.id.String() + ":" + nodeID))
	return "automata-" + hex.EncodeToString(sum[:])
}

// Lock the outbox row before every checkpoint mutation. A stale worker cannot
// overwrite progress committed under a replacement lease.
func (s *runNodeState) transaction(ctx context.Context, operation func(context.Context, pgx.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, outboxDatabaseTimeout)
	defer cancel()
	tx, err := s.repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", node.ErrStateUnavailable, err)
	}
	defer tx.Rollback(context.Background())
	var owned bool
	err = tx.QueryRow(ctx, `SELECT true FROM schedule_outbox
 WHERE run_id=$1 AND lease_token=$2 AND lease_until > clock_timestamp() FOR UPDATE`, s.claim.id, s.claim.token).Scan(&owned)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrLeaseLost
	}
	if err == nil {
		err = operation(ctx, tx)
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		return fmt.Errorf("%w: %w", node.ErrStateUnavailable, err)
	}
	return nil
}

func (s *runNodeState) Begin(ctx context.Context, nodeID string, maxAttempts int) (node.Attempt, error) {
	var attempt node.Attempt
	var exhausted error
	err := s.transaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var status string
		var output []byte
		var message *string
		var terminal bool
		var retryAt *time.Time
		err := tx.QueryRow(ctx, `SELECT status, attempts, output, error_message, terminal_failure, retry_at
 FROM workflow_node_runs WHERE run_id=$1 AND node_id=$2`, s.claim.id, nodeID).
			Scan(&status, &attempt.Number, &output, &message, &terminal, &retryAt)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if status == "succeeded" {
			attempt.Completed = true
			return json.Unmarshal(output, &attempt.Output)
		}
		if terminal && message != nil {
			exhausted = &node.RetryError{Message: *message}
			return nil
		}
		if attempt.Number >= maxAttempts {
			// A running attempt may have performed an external effect before a
			// crash. Never silently exceed the configured delivery budget.
			exhausted = fmt.Errorf("node %q exhausted %d attempts; last outcome may be unknown", nodeID, maxAttempts)
			if message != nil {
				exhausted = fmt.Errorf("node %q exhausted %d attempts: %s", nodeID, maxAttempts, *message)
			}
			return nil
		}
		if retryAt != nil && time.Now().Before(*retryAt) {
			exhausted = &node.RetryError{Message: "waiting to retry", Retryable: true, After: time.Until(*retryAt)}
			return nil
		}
		attempt.Number++
		_, err = tx.Exec(ctx, `INSERT INTO workflow_node_runs(run_id,node_id,status,attempts,started_at)
 VALUES($1,$2,'running',$3,clock_timestamp()) ON CONFLICT(run_id,node_id) DO UPDATE
 SET status='running', attempts=$3, error_message=NULL, updated_at=clock_timestamp()`, s.claim.id, nodeID, attempt.Number)
		return err
	})
	if err != nil {
		return attempt, err
	}
	return attempt, exhausted
}

func (s *runNodeState) Complete(ctx context.Context, nodeID string, output map[string]any) error {
	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("%w: encode output: %w", node.ErrStateUnavailable, err)
	}
	return s.transaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE workflow_node_runs SET status='succeeded', output=$3::jsonb,
 error_message=NULL, updated_at=clock_timestamp() WHERE run_id=$1 AND node_id=$2`, s.claim.id, nodeID, data)
		return err
	})
}

func (s *runNodeState) Fail(ctx context.Context, nodeID string, failure error) error {
	var classified *node.RetryError
	terminal := errors.As(failure, &classified) && !classified.Retryable
	var retryAt *time.Time
	if classified != nil && classified.After > 0 {
		at := time.Now().Add(classified.After)
		retryAt = &at
	}
	return s.transaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE workflow_node_runs SET status='failed', error_message=$3,
 terminal_failure=$4, retry_at=$5, updated_at=clock_timestamp() WHERE run_id=$1 AND node_id=$2`, s.claim.id, nodeID, failure.Error(), terminal, retryAt)
		return err
	})
}
