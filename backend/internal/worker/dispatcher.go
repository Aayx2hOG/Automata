package worker

import (
	"context"
	"time"

	"github.com/Aayx2hOG/automata/internal/queue"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReadyRuns interface {
	ListReadyRuns(context.Context, int) ([]uuid.UUID, error)
}

// Dispatch copies durable IDs into a bounded local queue. Full queues and
// restarts leave the authoritative outbox untouched for the next poll.
func Dispatch(ctx context.Context, runs ReadyRuns, q *queue.Queue, logger *zap.Logger) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for ctx.Err() == nil {
		ids, err := runs.ListReadyRuns(ctx, q.Capacity())
		if err != nil && ctx.Err() == nil {
			logger.Error("outbox dispatch failed", zap.Error(err))
		}
		for _, id := range ids {
			q.Enqueue(queue.Job{RunID: id})
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
