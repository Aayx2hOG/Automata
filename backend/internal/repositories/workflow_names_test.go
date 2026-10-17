package repositories

import (
	"context"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"sync"
	"testing"
)

func TestUnnamedWorkflowsReceiveDistinctNames(t *testing.T) {
	fixture, schedule := outboxFixture(t)
	ctx := context.Background()
	var owner uuid.UUID
	if err := fixture.pool.QueryRow(ctx, `SELECT owner_id FROM workflows WHERE id=$1`, schedule.WorkflowID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	repo := &pgWorkflowRepository{pool: fixture.pool}
	graph := models.WorkflowGraph{}
	named := &models.Workflow{OwnerID: owner, Name: "untitled_1", IsActive: true}
	if _, err := repo.CreateWithInitialVersion(ctx, named, graph); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	names := make(chan string, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &models.Workflow{OwnerID: owner, Name: "  ", IsActive: true}
			if _, err := repo.CreateWithInitialVersion(ctx, w, graph); err != nil {
				t.Error(err)
				return
			}
			names <- w.Name
		}()
	}
	wg.Wait()
	close(names)
	seen := map[string]bool{}
	for name := range names {
		if seen[name] {
			t.Errorf("duplicate name %s", name)
		}
		seen[name] = true
	}
	if !seen["untitled_2"] || !seen["untitled_3"] {
		t.Fatalf("unexpected names: %v", seen)
	}
}
