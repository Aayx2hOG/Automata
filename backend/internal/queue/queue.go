package queue

import (
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

type Job struct {
	RunID        uuid.UUID
	Graph        models.WorkflowGraph
	SeedNodeType string
	SeedData     map[string]any
}

type Queue struct {
	jobs chan Job
}

func NewQueue(bufferSize int) *Queue {
	return &Queue{jobs: make(chan Job, bufferSize)}
}

func (q *Queue) Enqueue(job Job) bool {
	select {
	case q.jobs <- job:
		return true
	default:
		return false
	}
}

func (q *Queue) Jobs() <-chan Job {
	return q.jobs
}

func (q *Queue) Close() {
	close(q.jobs)
}
