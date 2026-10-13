package workflow

import (
	"errors"
	"maps"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/node"
)

func (e *Engine) executeNode(ctx *node.ExecutionContext, definition models.GraphNode, outputs map[string]map[string]any, seed *seed) (map[string]any, error) {
	maxAttempts, backoff := 1, time.Duration(0)
	if definition.Retry != nil {
		maxAttempts = definition.Retry.MaxAttempts
		backoff = time.Duration(definition.Retry.BackoffSeconds) * time.Second
	}
	state := node.StateFromContext(ctx.Ctx)
	for attempt := 1; ; attempt++ {
		if err := ctx.Ctx.Err(); err != nil {
			return nil, err
		}
		key := ""
		if state != nil {
			stored, err := state.Begin(ctx.Ctx, definition.ID, maxAttempts)
			if err != nil {
				var retry *node.RetryError
				if errors.As(err, &retry) && retry.Retryable && retry.After > 0 {
					if err := waitForRetry(ctx, retry.After); err != nil {
						return nil, err
					}
					continue
				}
				return nil, err
			}
			if stored.Completed {
				return stored.Output, nil
			}
			attempt = stored.Number
			key = state.IdempotencyKey(definition.ID)
		}
		var output map[string]any
		var executionErr error
		if seed != nil && definition.Type == seed.nodeType {
			output = maps.Clone(seed.data)
		} else {
			impl, err := e.registry.Build(definition.Type)
			if err != nil {
				return nil, err
			}
			config, err := resolveConfig(definition, outputs)
			if err != nil {
				executionErr = &node.RetryError{Message: err.Error()}
			} else {
				output, executionErr = impl.Execute(&node.ExecutionContext{
					Ctx: ctx.Ctx, NodeID: definition.ID, Config: config,
					Outputs: outputs, IdempotencyKey: key,
				})
			}
		}
		if executionErr == nil {
			if state != nil {
				if err := state.Complete(ctx.Ctx, definition.ID, output); err != nil {
					return nil, err
				}
			}
			return output, nil
		}
		if err := ctx.Ctx.Err(); err != nil {
			return nil, err
		}
		if state != nil {
			if err := state.Fail(ctx.Ctx, definition.ID, executionErr); err != nil {
				return nil, err
			}
		}
		var classified *node.RetryError
		classifiedFailure := errors.As(executionErr, &classified)
		if attempt >= maxAttempts || (classifiedFailure && !classified.Retryable) {
			return nil, executionErr
		}
		wait := backoff
		if classifiedFailure && classified.After > wait {
			wait = classified.After
		}
		if err := waitForRetry(ctx, wait); err != nil {
			return nil, err
		}
	}
}

func waitForRetry(ctx *node.ExecutionContext, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Ctx.Done():
		return ctx.Ctx.Err()
	case <-timer.C:
		return nil
	}
}
