package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

func TestConcurrentAdmissionCannotExceedOwnerQuota(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	var version uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM workflow_versions WHERE workflow_id=$1`, schedule.WorkflowID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	limits := AdmissionLimits{MaxPending: 100, MaxPendingPerUser: 3, RunsPerMinute: 100}
	r.limits = limits
	var wg sync.WaitGroup
	results := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Independent repository instances share the same database quota.
			runs := NewWorkflowRunRepository(r.pool, limits)
			trigger := models.TriggerManual
			if i%2 == 0 {
				trigger = models.TriggerWebhook
			}
			_, err := runs.Enqueue(ctx, schedule.WorkflowID, version, trigger, nil)
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, models.ErrRunQuotaExceeded) {
			t.Fatal(err)
		}
	}
	if successes != 3 {
		t.Fatalf("accepted %d runs, want 3", successes)
	}
	assertCount(t, r, "workflow_runs", 3)
	assertCount(t, r, "schedule_outbox", 3)
	now := time.Now()
	if err := r.EnqueueRun(ctx, schedule, now, now.Add(time.Minute)); !errors.Is(err, models.ErrRunQuotaExceeded) {
		t.Fatalf("scheduler bypassed quota: %v", err)
	}
	due, err := r.ListDue(ctx, now)
	if err != nil || len(due) != 1 {
		t.Fatalf("rejected schedule lost: %v %v", due, err)
	}
}

func TestRollingQuotaIncludesFinishedRuns(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	r.limits = AdmissionLimits{MaxPending: 10, MaxPendingPerUser: 10, RunsPerMinute: 1}
	id := enqueueFixture(t, r, schedule)
	if _, err := r.ProcessRun(ctx, id, func(context.Context, models.WorkflowGraph, map[string]any) (map[string]any, error) { return nil, nil }); err != nil {
		t.Fatal(err)
	}
	var version uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM workflow_versions WHERE workflow_id=$1`, schedule.WorkflowID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	runs := NewWorkflowRunRepository(r.pool, r.limits)
	if _, err := runs.Enqueue(ctx, schedule.WorkflowID, version, models.TriggerManual, nil); !errors.Is(err, models.ErrRunQuotaExceeded) {
		t.Fatalf("finished run bypassed rolling quota: %v", err)
	}
	if _, err := r.pool.Exec(ctx, `UPDATE workflow_runs SET created_at=clock_timestamp()-interval '2 minutes'`); err != nil {
		t.Fatal(err)
	}
	if _, err := runs.Enqueue(ctx, schedule.WorkflowID, version, models.TriggerManual, nil); err != nil {
		t.Fatalf("expired quota did not recover: %v", err)
	}
}

func TestGlobalBacklogAdmission(t *testing.T) {
	r, schedule := outboxFixture(t)
	ctx := context.Background()
	r.limits = AdmissionLimits{MaxPending: 1, MaxPendingPerUser: 100, RunsPerMinute: 100}
	enqueueFixture(t, r, schedule)
	var version uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM workflow_versions WHERE workflow_id=$1`, schedule.WorkflowID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if _, err := NewWorkflowRunRepository(r.pool, r.limits).Enqueue(ctx, schedule.WorkflowID, version, models.TriggerManual, nil); !errors.Is(err, models.ErrQueueFull) {
		t.Fatalf("global backlog limit bypassed: %v", err)
	}
	assertCount(t, r, "workflow_runs", 1)
}
