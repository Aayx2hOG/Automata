package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func enqueueFixture(t *testing.T, r *pgScheduleRepository, s models.Schedule) uuid.UUID {
	t.Helper()
	now := time.Now()
	if err := r.EnqueueRun(context.Background(), s, now, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	var id uuid.UUID
	if err := r.pool.QueryRow(context.Background(), `SELECT run_id FROM schedule_outbox`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestDurableWebhookPayloadReachesEngine(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	var version uuid.UUID
	if err := r.pool.QueryRow(ctx, `UPDATE workflow_versions SET graph='{"nodes":[{"id":"hook","type":"webhook_trigger"}],"edges":[]}' WHERE workflow_id=$1 RETURNING id`, s.WorkflowID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	runs := &pgWorkflowRunRepository{pool: r.pool}
	run, err := runs.Enqueue(ctx, s.WorkflowID, version, models.TriggerWebhook, map[string]interface{}{"payload": "persisted"})
	if err != nil {
		t.Fatal(err)
	}
	engine := workflow.NewEngine(workflow.NewRegistry(zap.NewNop()))
	ok, err := r.ProcessRun(ctx, run.ID, func(ctx context.Context, graph models.WorkflowGraph, seed map[string]interface{}) (map[string]interface{}, error) {
		result := engine.RunWithSeededOutput(&node.ExecutionContext{Ctx: ctx}, graph, "webhook_trigger", seed)
		return map[string]interface{}{"hook": result.Outputs["hook"]}, result.Error
	})
	if err != nil || !ok {
		t.Fatalf("execution: %v %v", ok, err)
	}
	var payload string
	if err = r.pool.QueryRow(ctx, `SELECT outputs->'hook'->>'payload' FROM workflow_runs WHERE id=$1`, run.ID).Scan(&payload); err != nil || payload != "persisted" {
		t.Fatalf("persisted result: %s %v", payload, err)
	}
}

func TestExecutionReleasesConnectionAndPublishesRunningStatus(t *testing.T) {
	r, s := outboxFixture(t)
	id := enqueueFixture(t, r, s)
	cfg := r.pool.Config()
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	consumer := &pgScheduleRepository{pool: pool}
	ok, err := consumer.ProcessRun(context.Background(), id, func(ctx context.Context, _ models.WorkflowGraph, _ map[string]interface{}) (map[string]interface{}, error) {
		queryCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		var status string
		// With MaxConns=1 this blocks if execution still holds a connection.
		if err := pool.QueryRow(queryCtx, `SELECT status FROM workflow_runs WHERE id=$1`, id).Scan(&status); err != nil {
			return nil, err
		}
		if status != "running" {
			t.Errorf("visible status = %s", status)
		}
		return nil, nil
	})
	if err != nil || !ok {
		t.Fatalf("execution: %v %v", ok, err)
	}
	var status string
	if err = pool.QueryRow(context.Background(), `SELECT status FROM workflow_runs WHERE id=$1`, id).Scan(&status); err != nil || status != "succeeded" {
		t.Fatalf("status: %s %v", status, err)
	}
}

func TestExpiredLeaseCannotRenewAcknowledgeOrReleaseNewOwner(t *testing.T) {
	r, s := outboxFixture(t)
	id := enqueueFixture(t, r, s)
	ctx := context.Background()
	old, err := r.claimRun(ctx, id)
	if err != nil || old == nil {
		t.Fatalf("claim: %v", err)
	}
	if _, err = r.pool.Exec(ctx, `UPDATE schedule_outbox SET lease_until=clock_timestamp()-interval '1 second' WHERE run_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err = r.finishRun(ctx, old, nil, nil); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("expired acknowledgement: %v", err)
	}
	next, err := r.claimRun(ctx, id)
	if err != nil || next == nil {
		t.Fatalf("reclaim: %v", err)
	}
	if err = r.renewRun(ctx, old); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale renewal: %v", err)
	}
	if err = r.finishRun(ctx, old, nil, nil); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale acknowledgement: %v", err)
	}
	r.releaseRun(old)
	if err = r.finishRun(ctx, next, map[string]interface{}{"owner": "new"}, nil); err != nil {
		t.Fatal(err)
	}
	assertCount(t, r, "schedule_outbox", 0)
}

func TestHeartbeatKeepsLongRunningJobClaimed(t *testing.T) {
	r, s := outboxFixture(t)
	id := enqueueFixture(t, r, s)
	r.leaseDuration = 600 * time.Millisecond
	ok, err := r.ProcessRun(context.Background(), id, func(ctx context.Context, _ models.WorkflowGraph, _ map[string]interface{}) (map[string]interface{}, error) {
		select {
		case <-time.After(1200 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		other, err := r.claimRun(ctx, id)
		if err != nil {
			return nil, err
		}
		if other != nil {
			t.Error("running job lost its lease despite heartbeats")
		}
		return nil, nil
	})
	if err != nil || !ok {
		t.Fatalf("execution: %v %v", ok, err)
	}
}

func TestLeaseLossCancelsExecutorAndFencesResult(t *testing.T) {
	r, s := outboxFixture(t)
	id := enqueueFixture(t, r, s)
	r.leaseDuration = 600 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var replacement *claimedRun
	ok, err := r.ProcessRun(ctx, id, func(execCtx context.Context, _ models.WorkflowGraph, _ map[string]interface{}) (map[string]interface{}, error) {
		if _, err := r.pool.Exec(ctx, `UPDATE schedule_outbox SET lease_until=clock_timestamp()-interval '1 second' WHERE run_id=$1`, id); err != nil {
			return nil, err
		}
		var err error
		replacement, err = r.claimRun(ctx, id)
		if err != nil {
			return nil, err
		}
		<-execCtx.Done()
		return map[string]interface{}{"stale": true}, nil
	})
	if ok || !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale worker: %v %v", ok, err)
	}
	if replacement == nil {
		t.Fatal("replacement did not claim")
	}
	if err := r.finishRun(ctx, replacement, nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestAcknowledgementFailureKeepsDurableWork(t *testing.T) {
	r, s := outboxFixture(t)
	id := enqueueFixture(t, r, s)
	ctx := context.Background()
	if _, err := r.pool.Exec(ctx, `ALTER TABLE workflow_runs ADD CONSTRAINT reject_success CHECK (status <> 'succeeded')`); err != nil {
		t.Fatal(err)
	}
	ok, err := r.ProcessRun(ctx, id, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	if ok || err == nil {
		t.Fatalf("expected failed acknowledgement: %v %v", ok, err)
	}
	assertCount(t, r, "schedule_outbox", 1)
	if _, err = r.pool.Exec(ctx, `ALTER TABLE workflow_runs DROP CONSTRAINT reject_success`); err != nil {
		t.Fatal(err)
	}
	ok, err = r.ProcessRun(ctx, id, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	if !ok || err != nil {
		t.Fatalf("retry: %v %v", ok, err)
	}
}

func TestWorkflowAndInitialVersionRollbackTogether(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	var owner uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT owner_id FROM workflows WHERE id=$1`, s.WorkflowID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	repo := &pgWorkflowRepository{pool: r.pool}
	if _, err := r.pool.Exec(ctx, `ALTER TABLE workflow_versions ADD CONSTRAINT reject_version CHECK (false) NOT VALID`); err != nil {
		t.Fatal(err)
	}
	w := &models.Workflow{OwnerID: owner, Name: "atomic", IsActive: true}
	graph := models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "root", Type: "manual_trigger"}}}
	if _, err := repo.CreateWithInitialVersion(ctx, w, graph); err == nil {
		t.Fatal("expected failed initial version")
	}
	if w.ID != uuid.Nil {
		t.Fatal("failed creation published a workflow ID")
	}
	assertCount(t, r, "workflows", 1)
	assertCount(t, r, "workflow_versions", 1)
	if _, err := r.pool.Exec(ctx, `ALTER TABLE workflow_versions DROP CONSTRAINT reject_version`); err != nil {
		t.Fatal(err)
	}
	version, err := repo.CreateWithInitialVersion(ctx, w, graph)
	if err != nil || version.WorkflowID != w.ID || version.Version != 1 {
		t.Fatalf("create: %+v %v", version, err)
	}
	assertCount(t, r, "workflows", 2)
	assertCount(t, r, "workflow_versions", 2)
}
