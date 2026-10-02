package worker

import (
	"context"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type readyFixture struct{ ids []uuid.UUID }

func (f readyFixture) ListReadyRuns(context.Context, int) ([]uuid.UUID, error) { return f.ids, nil }

func TestDispatcherAndPoolBoundConcurrencyAndStop(t *testing.T) {
	q := queue.NewQueue(8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan uuid.UUID, 3)
	release := make(chan struct{})
	pool := NewPool(q, func(ctx context.Context, job queue.Job) {
		started <- job.RunID
		select {
		case <-release:
		case <-ctx.Done():
		}
	}, 2, zap.NewNop())
	pool.Start(ctx)
	dispatched := make(chan struct{})
	go func() {
		defer close(dispatched)
		Dispatch(ctx, readyFixture{[]uuid.UUID{uuid.New(), uuid.New(), uuid.New()}}, q, zap.NewNop())
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("workers did not execute dispatched jobs")
		}
	}
	select {
	case <-started:
		t.Fatal("pool exceeded concurrency limit")
	case <-time.After(30 * time.Millisecond):
	}
	release <- struct{}{}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("queued job never started")
	}
	cancel()
	q.Close()
	shutdown, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	if err := pool.Shutdown(shutdown); err != nil {
		t.Fatal(err)
	}
	select {
	case <-dispatched:
	case <-shutdown.Done():
		t.Fatal("dispatcher did not stop")
	}
}

func TestWorkerSurvivesPanicAndReleasesHint(t *testing.T) {
	q := queue.NewQueue(2)
	first := queue.Job{RunID: uuid.New()}
	second := queue.Job{RunID: uuid.New()}
	q.Enqueue(first)
	q.Enqueue(second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	pool := NewPool(q, func(_ context.Context, job queue.Job) {
		if job.RunID == first.RunID {
			panic("broken node")
		}
		close(done)
	}, 1, zap.NewNop())
	pool.Start(ctx)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker stopped after panic")
	}
	cancel()
	shutdown, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := pool.Shutdown(shutdown); err != nil {
		t.Fatal(err)
	}
	if !q.Enqueue(first) {
		t.Fatal("panic left job permanently deduplicated")
	}
	q.Close()
}
