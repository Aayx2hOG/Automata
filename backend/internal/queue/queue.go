package queue

import (
	"github.com/google/uuid"
	"sync"
)

type Job struct {
	RunID uuid.UUID
}

type Queue struct {
	jobs    chan Job
	mu      sync.Mutex
	pending map[uuid.UUID]bool
	closed  bool
}

func NewQueue(bufferSize int) *Queue {
	return &Queue{jobs: make(chan Job, bufferSize), pending: make(map[uuid.UUID]bool)}
}

func (q *Queue) Enqueue(job Job) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed || q.pending[job.RunID] {
		return false
	}
	select {
	case q.jobs <- job:
		q.pending[job.RunID] = true
		return true
	default:
		return false
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.closed {
		q.closed = true
		close(q.jobs)
	}
}

func (q *Queue) Capacity() int { return cap(q.jobs) }

func (q *Queue) Done(job Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.pending, job.RunID)
}
