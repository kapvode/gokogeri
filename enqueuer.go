package gokogeri

import (
	"context"
	"fmt"

	"github.com/kapvode/gokogeri/internal/redisutil"
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
		return fmt.Errorf("encode job: %v", err)
	}

	conn, err := e.cp.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get conn: %v", err)
	}
	defer conn.Close()

	err = conn.Send("SADD", "queues", j.enc.Queue)
	if err != nil {
		return fmt.Errorf("send: %v", err)
	}

	err = conn.Send("LPUSH", "queue:"+j.enc.Queue, enc)
	if err != nil {
		return fmt.Errorf("send: %v", err)
	}

	_, err = redisutil.DoMany(conn, 2)
	if err != nil {
		return fmt.Errorf("enqueue job: %v", err)
	}

	return nil
}
