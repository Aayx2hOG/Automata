package queue

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

func TestBoundedQueueDeduplicatesUntilJobFinishes(t *testing.T) {
	q := NewQueue(1)
	job := Job{RunID: uuid.New()}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if q.Enqueue(job) {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("duplicate dispatches accepted: %d", accepted.Load())
	}
	if q.Enqueue(Job{RunID: uuid.New()}) {
		t.Fatal("full queue accepted work")
	}
	got := <-q.Jobs()
	if q.Enqueue(job) {
		t.Fatal("executing job was enqueued again")
	}
	q.Done(got)
	if !q.Enqueue(job) {
		t.Fatal("released hint could not be retried")
	}
	q.Close()
	q.Close()
	if q.Enqueue(Job{RunID: uuid.New()}) {
		t.Fatal("closed queue accepted work")
	}
}

func TestCloseAndEnqueueAreSafeConcurrently(t *testing.T) {
	q := NewQueue(8)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); q.Enqueue(Job{RunID: uuid.New()}); q.Close() }()
	}
	wg.Wait()
}
