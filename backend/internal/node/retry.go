package node

import "time"

// RetryError classifies an execution failure without exposing response bodies.
type RetryError struct {
	Message   string
	Retryable bool
	After     time.Duration
}

func (e *RetryError) Error() string { return e.Message }
