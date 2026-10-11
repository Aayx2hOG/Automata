package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"go.uber.org/zap"
)

func TestRecoveryRestoresCompletedNodesAndAttemptBudget(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	graph := `{"nodes":[{"id":"done","type":"manual_trigger"},{"id":"interrupted","type":"manual_trigger","retry":{"max_attempts":2,"backoff_seconds":0}}],"edges":[{"from_node_id":"done","to_node_id":"interrupted"}]}`
	if _, err := r.pool.Exec(ctx, `UPDATE workflow_versions SET graph=$1::jsonb`, graph); err != nil {
		t.Fatal(err)
	}
	id := enqueueFixture(t, r, schedule)
	interrupted, cancel := context.WithCancel(ctx)
	var key string
	ok, err := r.ProcessRun(interrupted, id, func(ctx context.Context, _ models.WorkflowGraph, _ map[string]any) (map[string]any, error) {
		state := node.StateFromContext(ctx)
		if _, err := state.Begin(ctx, "done", 1); err != nil {
			t.Fatal(err)
		}
		if err := state.Complete(ctx, "done", map[string]any{"checkpoint": "original-output"}); err != nil {
			t.Fatal(err)
		}
		if _, err := state.Begin(ctx, "interrupted", 2); err != nil {
			t.Fatal(err)
		}
		key = state.IdempotencyKey("interrupted")
		cancel() // Simulate loss after an attempt started, before its checkpoint.
		return nil, context.Canceled
	})
	if ok || !errors.Is(err, context.Canceled) {
		t.Fatalf("interruption: %v %v", ok, err)
	}
	consumer := &pgScheduleRepository{pool: r.pool}
	engine := workflow.NewEngine(workflow.NewRegistry(zap.NewNop()))
	ok, err = consumer.ProcessRun(ctx, id, func(ctx context.Context, graph models.WorkflowGraph, _ map[string]any) (map[string]any, error) {
		if node.StateFromContext(ctx).IdempotencyKey("interrupted") != key {
			t.Fatal("idempotency key changed on recovery")
		}
		result := engine.Run(&node.ExecutionContext{Ctx: ctx}, graph)
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		if result.Outputs["done"]["checkpoint"] != "original-output" {
			t.Fatal("completed node executed again")
		}
		return nil, result.Error
	})
	if !ok || err != nil {
		t.Fatalf("recovery: %v %v", ok, err)
	}
	var first, second int
	if err := r.pool.QueryRow(ctx, `SELECT attempts FROM workflow_node_runs WHERE run_id=$1 AND node_id='done'`, id).Scan(&first); err != nil {
		t.Fatal(err)
	}
	if err := r.pool.QueryRow(ctx, `SELECT attempts FROM workflow_node_runs WHERE run_id=$1 AND node_id='interrupted'`, id).Scan(&second); err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("attempts = %d,%d", first, second)
	}
}

func TestNodeStateFencesStaleWorkersAndDoesNotResetAttempts(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	id := enqueueFixture(t, r, schedule)
	claim, err := r.claimRun(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	old := &runNodeState{repository: r, claim: claim}
	if _, err := old.Begin(ctx, "write", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := r.pool.Exec(ctx, `UPDATE schedule_outbox SET lease_until=clock_timestamp()-interval '1 second' WHERE run_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	replacement, err := r.claimRun(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := old.Complete(ctx, "write", map[string]any{"stale": true}); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale checkpoint accepted: %v", err)
	}
	current := &runNodeState{repository: r, claim: replacement}
	if _, err := current.Begin(ctx, "write", 1); err == nil || errors.Is(err, node.ErrStateUnavailable) {
		t.Fatalf("attempt budget reset or treated as transient: %v", err)
	}
}

func TestCheckpointFailurePreservesOutbox(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	id := enqueueFixture(t, r, schedule)
	ok, err := r.ProcessRun(ctx, id, func(ctx context.Context, _ models.WorkflowGraph, _ map[string]any) (map[string]any, error) {
		state := node.StateFromContext(ctx)
		if _, err := state.Begin(ctx, "write", 1); err != nil {
			return nil, err
		}
		if _, err := r.pool.Exec(ctx, `ALTER TABLE workflow_node_runs ADD CONSTRAINT reject_checkpoint CHECK (status <> 'succeeded')`); err != nil {
			t.Fatal(err)
		}
		return nil, state.Complete(ctx, "write", map[string]any{"ok": true})
	})
	if ok || !errors.Is(err, node.ErrStateUnavailable) {
		t.Fatalf("checkpoint failure: %v %v", ok, err)
	}
	assertCount(t, r, "schedule_outbox", 1)
}
