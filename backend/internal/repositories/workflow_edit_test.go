package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

func TestWorkflowEditsPreserveIdentityAndIsolateOwners(t *testing.T) {
	fixture, schedule := outboxFixture(t)
	ctx := context.Background()
	workflows := NewWorkflowRepository(fixture.pool)
	versions := NewWorkflowVersionRepository(fixture.pool)
	runs := NewWorkflowRunRepository(fixture.pool)
	original, err := workflows.GetById(ctx, schedule.WorkflowID)
	if err != nil {
		t.Fatal(err)
	}
	graph := models.WorkflowGraph{Nodes: []models.GraphNode{{ID: "start", Type: "manual_trigger", Label: "Start here", Position: &models.NodePosition{X: 120, Y: 240}}}, Edges: []models.GraphEdge{}}
	if _, err := workflows.UpdateWithVersion(ctx, original.ID, uuid.New(), "forbidden", "", graph); !errors.Is(err, models.ErrWorkflowNotFound) {
		t.Fatalf("foreign update: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := workflows.UpdateWithVersion(ctx, original.ID, original.OwnerID, "Edited", "saved", graph); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	saved, err := workflows.GetById(ctx, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Name != "Edited" || saved.WebhookSecret != original.WebhookSecret {
		t.Fatalf("identity changed: %+v", saved)
	}
	latest, err := versions.GetLatestByWorkflow(ctx, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Version != 4 || latest.Graph.Nodes[0].Label != "Start here" || latest.Graph.Nodes[0].Position.X != 120 {
		t.Fatalf("bad saved graph: %+v", latest)
	}
	run, err := runs.Enqueue(ctx, original.ID, latest.ID, models.TriggerManual, nil)
	if err != nil {
		t.Fatal(err)
	}
	own, err := runs.ListByOwner(ctx, original.OwnerID, 50, 0)
	if err != nil || len(own) != 1 || own[0].ID != run.ID || own[0].WorkflowVersionID != latest.ID {
		t.Fatalf("history: %+v %v", own, err)
	}
	other, err := runs.ListByOwner(ctx, uuid.New(), 50, 0)
	if err != nil || len(other) != 0 {
		t.Fatalf("foreign history: %+v %v", other, err)
	}
	page, err := runs.ListByOwner(ctx, original.OwnerID, 1, 1)
	if err != nil || len(page) != 0 {
		t.Fatalf("pagination: %+v %v", page, err)
	}
}
