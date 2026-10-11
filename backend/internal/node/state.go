package node

import (
	"context"
	"errors"
)

// ErrStateUnavailable leaves the durable job pending instead of recording a
// terminal workflow failure when its checkpoint cannot be read or committed.
var ErrStateUnavailable = errors.New("node execution state unavailable")

type Attempt struct {
	Number    int
	Completed bool
	Output    map[string]any
}

// State persists attempts and outputs under the current run's execution lease.
// Begin returns a completed output without consuming another attempt.
type State interface {
	Begin(context.Context, string, int) (Attempt, error)
	Complete(context.Context, string, map[string]any) error
	Fail(context.Context, string, error) error
	IdempotencyKey(string) string
}

type stateKey struct{}

func WithState(ctx context.Context, state State) context.Context {
	return context.WithValue(ctx, stateKey{}, state)
}

func StateFromContext(ctx context.Context) State {
	state, _ := ctx.Value(stateKey{}).(State)
	return state
}
