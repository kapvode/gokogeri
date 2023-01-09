package gokogeri

import (
	"context"
	"fmt"
)

// Enqueuer puts jobs in queues.
type Enqueuer struct {
	cp ConnProvider
}

// NewEnqueuer returns a new instance.
func NewEnqueuer(cp ConnProvider) *Enqueuer {
	return &Enqueuer{cp: cp}
}

// Enqueue adds the job to the queue configured in the job, or the default one, if no queue is configured.
func (e *Enqueuer) Enqueue(ctx context.Context, j *Job) error {
	err := j.setDefaults()
	if err != nil {
		return fmt.Errorf("setting job defaults: %v", err)
	}

	enc, err := j.encode()
	if err != nil {
		return fmt.Errorf("encode job: %w", err)
	}

	conn, err := e.cp.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}
	defer conn.Close()

	_, err = conn.Do("SADD", "queues", j.enc.Queue)

	_, err = conn.Do("LPUSH", "queue:"+j.enc.Queue, enc)
	if err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}

	return nil
}
