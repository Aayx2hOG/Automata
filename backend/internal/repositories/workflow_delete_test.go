package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

func TestWorkflowDeleteEnforcesOwnershipAndCascades(t *testing.T) {
	fixture, schedule := outboxFixture(t)
	ctx := context.Background()
	repo := NewWorkflowRepository(fixture.pool)
	workflow, err := repo.GetById(ctx, schedule.WorkflowID)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, workflow.ID, uuid.New()); !errors.Is(err, models.ErrWorkflowNotFound) {
		t.Fatalf("foreign delete: %v", err)
	}
	assertCount(t, fixture, "workflows", 1)
	// Completed history is deleted along with versions and schedules.
	_, err = fixture.pool.Exec(ctx, `INSERT INTO workflow_runs (workflow_id, workflow_version_id, trigger_type, status) SELECT workflow_id,id,'manual','succeeded' FROM workflow_versions WHERE workflow_id=$1`, workflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, workflow.ID, workflow.OwnerID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"workflows", "workflow_versions", "workflow_runs", "schedules", "schedule_outbox"} {
		assertCount(t, fixture, table, 0)
	}
	if err = repo.Delete(ctx, workflow.ID, workflow.OwnerID); !errors.Is(err, models.ErrWorkflowNotFound) {
		t.Fatalf("repeat delete: %v", err)
	}
}

func TestWorkflowDeletePreservesQueuedAndRunningWork(t *testing.T) {
	fixture, schedule := outboxFixture(t)
	ctx := context.Background()
	repo := NewWorkflowRepository(fixture.pool)
	workflow, err := repo.GetById(ctx, schedule.WorkflowID)
	if err != nil {
		t.Fatal(err)
	}
	version, err := NewWorkflowVersionRepository(fixture.pool).GetLatestByWorkflow(ctx, workflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := NewWorkflowRunRepository(fixture.pool).Enqueue(ctx, workflow.ID, version.ID, models.TriggerManual, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"pending", "running", "succeeded"} {
		// Even a finished run cannot be deleted until its outbox entry is acknowledged.
		if _, err = fixture.pool.Exec(ctx, `UPDATE workflow_runs SET status=$2 WHERE id=$1`, run.ID, status); err != nil {
			t.Fatal(err)
		}
		if err = repo.Delete(ctx, workflow.ID, workflow.OwnerID); !errors.Is(err, models.ErrWorkflowBusy) {
			t.Fatalf("delete with %s durable work: %v", status, err)
		}
		assertCount(t, fixture, "workflows", 1)
		assertCount(t, fixture, "schedule_outbox", 1)
	}
}
