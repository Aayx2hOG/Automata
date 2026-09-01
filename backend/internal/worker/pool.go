package worker

import (
	"context"
	"sync"

	"github.com/Aayx2hOG/automata/internal/queue"
	"go.uber.org/zap"
)

type Handler func(ctx context.Context, job queue.Job)

type Pool struct {
	queue   *queue.Queue
	handler Handler
	size    int
	logger  *zap.Logger
	wg      sync.WaitGroup
}

func NewPool(q *queue.Queue, handler Handler, size int, logger *zap.Logger) *Pool {
	return &Pool{queue: q, handler: handler, size: size, logger: logger}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i+1)
	}
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	p.logger.Info("worker started", zap.Int("worker_id", id))

	for {
		select {
		case job, ok := <-p.queue.Jobs():
			if !ok {
				p.logger.Info("worker stopping: queue closed", zap.Int("worker_id", id))
				return
			}
			p.handler(ctx, job)
		case <-ctx.Done():
			p.logger.Info("worker stopping: context cancelled", zap.Int("worker_id", id))
			return
		}
	}
}

func (p *Pool) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
