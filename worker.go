package gokogeri

import (
	"context"
)

// A Worker processes jobs from one or more queues.
type Worker interface {
	// Work is safe for concurrent use.
	Work(context.Context, *Job) error
}

// WorkerFunc is an adapter to allow the use of functions as Workers.
type WorkerFunc func(context.Context, *Job) error

// Work implements Worker by delegating to the wrapped function.
func (w WorkerFunc) Work(ctx context.Context, j *Job) error {
	return w(ctx, j)
}
