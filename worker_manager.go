package gokogeri

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// workerManager controls a group of workers processing a set of queues.
type workerManager struct {
	log zerolog.Logger

	dq        *dequeuer
	worker    Worker
	instances int
}

func newWorkerManager(
	log zerolog.Logger,
	dqf *dequeuerFactory,
	qset QueueSet,
	worker Worker,
	instances int,
) *workerManager {
	return &workerManager{
		dq:        dqf.newDequeuer(qset),
		worker:    worker,
		instances: instances,
		log:       log.With().Str("component", "manager").Strs("queue_set", qset.Names()).Logger(),
	}
}

// Run starts the worker instances and blocks. Call Stop to initiate shutdown, which will ultimately unblock Run.
// The provided Context becomes the base context for the workers.
func (m *workerManager) Run(ctx context.Context) {
	m.log.Debug().Msg("Starting")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		m.dq.Run()
	}()

	wg.Add(m.instances)

	for i := 0; i < m.instances; i++ {
		n := i + 1
		go func() {
			defer wg.Done()
			m.jobLoop(ctx, n)
		}()
	}

	wg.Wait()
	m.log.Debug().Msg("Stopped")
}

// Stop initiates worker shutdown.
// It stops the process of getting new jobs from the queues.
//
// The jobs that are currently in progress will be allowed to finish, at which point the call to Run will return. The
// jobs might potentially be interrupted by cancelling the Context provided to Run.
//
// Stop does not block.
func (m *workerManager) Stop() {
	m.log.Debug().Msg("Stopping")
	m.dq.Stop()
}

func (m *workerManager) jobLoop(ctx context.Context, n int) {
	log := m.log.With().Int("worker", n).Logger()
	log.Debug().Msg("Running")

	defer log.Debug().Msg("Stopped")

	for {
		log.Trace().Msg("Waiting for a job")
		r, ok := <-m.dq.C
		if !ok {
			log.Debug().Msg("No more work")
			return
		}

		job, err := newJobFromJSON(r.P)
		if err != nil {
			log.Warn().Msg("Invalid job")
			continue
		}

		log = log.With().Str("job_id", job.ID()).Logger()
		log.Info().Msg("Processing")

		err = m.safelyWork(ctx, job)
		if err != nil {
			log.Warn().Msg("Job has failed")
		} else {
			log.Info().Msg("Job done")
		}
	}
}

func (m *workerManager) safelyWork(ctx context.Context, job *Job) (err error) {
	defer func() {
		if val := recover(); val != nil {
			err = fmt.Errorf("worker panic: %v", val)
		}
	}()
	return m.worker.Work(ctx, job)
}
