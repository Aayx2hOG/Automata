package repositories

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Each test uses an isolated schema. TEST_DATABASE_URL must point to a test DB.
func outboxFixture(t *testing.T) (*pgScheduleRepository, models.Schedule) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "outbox_test_" + uuid.New().String()[:8]
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); admin.Close() })
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, name := range []string{"000001_create_users", "000005_create_workflows", "000006_create_workflow_versions", "000007_create_workflow_runs", "000008_create_schedules", "000009_schedule_outbox", "000010_execution_security", "000011_outbox_leases"} {
		sql, err := os.ReadFile(filepath.Join("..", "database", "migrations", name+".up.sql"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			t.Fatal(err)
		}
	}
	var owner, workflowID uuid.UUID
	if err = pool.QueryRow(ctx, `INSERT INTO users(email,password_hash) VALUES ('test@example.com','test') RETURNING id`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workflows(owner_id,name) VALUES ($1,'test') RETURNING id`, owner).Scan(&workflowID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workflow_versions(workflow_id,version,graph) VALUES ($1,1,'{"nodes":[],"edges":[]}')`, workflowID); err != nil {
		t.Fatal(err)
	}
	repo := &pgScheduleRepository{pool: pool}
	schedule := models.Schedule{WorkflowID: workflowID, CronExpression: "* * * * *", IsActive: true, NextRunAt: time.Now().UTC().Truncate(time.Second).Add(-time.Minute)}
	if err = repo.Create(ctx, &schedule); err != nil {
		t.Fatal(err)
	}
	return repo, schedule
}

func assertCount(t *testing.T, r *pgScheduleRepository, table string, want int) {
	t.Helper()
	var n int
	if err := r.pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != want {
		t.Fatalf("%s count = %d, want %d", table, n, want)
	}
}

func TestScheduleEnqueueRollbackAndConcurrentRetry(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	now := time.Now()
	next := now.Add(time.Minute)
	// Fail after creating the execution but before advancing the schedule.
	if _, err := r.pool.Exec(ctx, `ALTER TABLE schedule_outbox ADD CONSTRAINT reject_insert CHECK (false)`); err != nil {
		t.Fatal(err)
	}
	if err := r.EnqueueRun(ctx, s, now, next); err == nil {
		t.Fatal("expected insert failure")
	}
	assertCount(t, r, "workflow_runs", 0)
	due, err := r.ListDue(ctx, now)
	if err != nil || len(due) != 1 || !due[0].NextRunAt.Equal(s.NextRunAt) {
		t.Fatalf("occurrence lost: %v %v", due, err)
	}
	if _, err = r.pool.Exec(ctx, `ALTER TABLE schedule_outbox DROP CONSTRAINT reject_insert`); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.EnqueueRun(ctx, s, now, next); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	assertCount(t, r, "workflow_runs", 1)
	assertCount(t, r, "schedule_outbox", 1)
	due, err = r.ListDue(ctx, now)
	if err != nil || len(due) != 0 {
		t.Fatalf("schedule not advanced: %v %v", due, err)
	}
}

func TestOutboxInterruptedConsumerAndAcknowledgement(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	now := time.Now()
	if err := r.EnqueueRun(ctx, s, now, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	interrupted, cancel := context.WithCancel(ctx)
	processed, err := r.ProcessNextRun(interrupted, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
		// Another consumer must skip the leased occurrence.
		other, err := r.ProcessNextRun(ctx, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
			t.Error("duplicate concurrent execution")
			return nil, nil
		})
		if err != nil || other {
			t.Errorf("competing consumer: %v %v", other, err)
		}
		cancel()
		return nil, context.Canceled
	})
	if processed || !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted consume: %v %v", processed, err)
	}
	assertCount(t, r, "schedule_outbox", 1)
	processed, err = r.ProcessNextRun(ctx, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"recovered": true}, nil
	})
	if err != nil || !processed {
		t.Fatalf("recovery: %v %v", processed, err)
	}
	assertCount(t, r, "workflow_runs", 1)
	assertCount(t, r, "schedule_outbox", 0)
	var status string
	if err = r.pool.QueryRow(ctx, `SELECT status FROM workflow_runs`).Scan(&status); err != nil || status != "succeeded" {
		t.Fatalf("result: %s %v", status, err)
	}
}

func TestOutboxRecordsTerminalFailure(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	now := time.Now()
	if err := r.EnqueueRun(ctx, s, now, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	processed, err := r.ProcessNextRun(ctx, func(context.Context, models.WorkflowGraph, map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("node failed")
	})
	if err != nil || !processed {
		t.Fatalf("consume: %v %v", processed, err)
	}
	assertCount(t, r, "schedule_outbox", 0)
	var status, message string
	if err = r.pool.QueryRow(ctx, `SELECT status,error_message FROM workflow_runs`).Scan(&status, &message); err != nil || status != "failed" || message != "node failed" {
		t.Fatalf("result: %s %s %v", status, message, err)
	}
}

func TestAllTriggersDurable(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	runs := &pgWorkflowRunRepository{pool: r.pool}
	var version uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT id FROM workflow_versions WHERE workflow_id=$1`, s.WorkflowID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	for _, trigger := range []models.TriggerType{models.TriggerManual, models.TriggerWebhook} {
		var seed map[string]interface{}
		if trigger == models.TriggerWebhook {
			seed = map[string]interface{}{"persisted": "payload"}
		}
		if _, err := runs.Enqueue(ctx, s.WorkflowID, version, trigger, seed); err != nil {
			t.Fatal(err)
		}
		// A fresh repository represents a consumer after process restart.
		consumer := &pgScheduleRepository{pool: r.pool}
		ok, err := consumer.ProcessNextRun(ctx, func(_ context.Context, _ models.WorkflowGraph, got map[string]interface{}) (map[string]interface{}, error) {
			if trigger == models.TriggerWebhook && got["persisted"] != "payload" {
				t.Fatalf("lost payload: %v", got)
			}
			if trigger == models.TriggerManual && got != nil {
				t.Fatalf("unexpected seed: %v", got)
			}
			return nil, nil
		})
		if err != nil || !ok {
			t.Fatalf("consume: %v %v", ok, err)
		}
	}
	if _, err := r.pool.Exec(ctx, `ALTER TABLE schedule_outbox ADD CONSTRAINT reject_insert CHECK(false)`); err != nil {
		t.Fatal(err)
	}
	if _, err := runs.Enqueue(ctx, s.WorkflowID, version, models.TriggerManual, nil); err == nil {
		t.Fatal("expected failure")
	}
	assertCount(t, r, "workflow_runs", 2)
}

func TestRefreshRotationConcurrencyAndInactiveUser(t *testing.T) {
	r, s := outboxFixture(t)
	ctx := context.Background()
	sql, err := os.ReadFile(filepath.Join("..", "database", "migrations", "000002_create_refresh_tokens.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.pool.Exec(ctx, string(sql)); err != nil {
		t.Fatal(err)
	}
	var owner uuid.UUID
	if err = r.pool.QueryRow(ctx, `SELECT owner_id FROM workflows WHERE id=$1`, s.WorkflowID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	tokens := &pgRefreshTokenRepository{pool: r.pool}
	expires := time.Now().Add(time.Hour)
	if err = tokens.Create(ctx, owner, "old", expires); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func() { results <- tokens.Rotate(ctx, "old", uuid.NewString(), owner, expires) }()
	}
	successes := 0
	for i := 0; i < 8; i++ {
		err := <-results
		if err == nil {
			successes++
		} else if !errors.Is(err, models.ErrorInvalidToken) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful rotations: %d", successes)
	}
	if err = tokens.Create(ctx, owner, "inactive", expires); err != nil {
		t.Fatal(err)
	}
	if _, err = r.pool.Exec(ctx, `UPDATE users SET is_active=false WHERE id=$1`, owner); err != nil {
		t.Fatal(err)
	}
	if err = tokens.Rotate(ctx, "inactive", "replacement", owner, expires); !errors.Is(err, models.ErrorInvalidToken) {
		t.Fatalf("inactive rotation: %v", err)
	}
}
