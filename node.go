package gokogeri

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// A Node represents a single server instance processing as many queues with as many Worker instances as are needed.
type Node struct {
	log    zerolog.Logger
	rawLog zerolog.Logger

	dqf *dequeuerFactory

	wg       sync.WaitGroup
	managers []*workerManager

	ctx    context.Context
	cancel context.CancelFunc
}

// NewNode returns a new instance.
func NewNode(log zerolog.Logger, cp ConnProvider, longPollTimeout int) *Node {
	n := &Node{
		dqf:    newDequeuerFactory(log, cp, longPollTimeout),
		log:    log.With().Str("component", "node").Logger(),
		rawLog: log,
	}
	n.ctx, n.cancel = context.WithCancel(context.Background())
	return n
}

// ProcessQueues configures the Node to process the given set of queues using the desired number of Worker instances.
//
// You can call ProcessQueues many times with different sets of queues and Workers.
// Do not call it any more after calling Run.
func (n *Node) ProcessQueues(qs QueueSet, w Worker, instances int) {
	n.managers = append(n.managers, newWorkerManager(n.rawLog, n.dqf, qs, w, instances))
}

// Run starts the process of getting jobs from queues and passing them to Workers.
// It blocks until the Node is shut down. See Stop for more.
func (n *Node) Run() {
	n.log.Debug().Msg("Starting managers")

	n.wg.Add(len(n.managers))
	for _, m := range n.managers {
		m := m
		go func() {
			defer n.wg.Done()
			m.Run(n.ctx)
		}()
	}

	n.log.Info().Msg("Running")
	n.wg.Wait()
}

// Stop initiates worker shutdown. Once the shutdown process is complete, the call to Run will return.
//
// New jobs will not be taken from queues any more.
//
// Workers that are currently processing jobs will be given a grace period and allowed to finish. Once the Context
// provided to Stop expires, the Context passed to every Worker will be cancelled.
//
// Stop blocks until the shutdown process has completed.
func (n *Node) Stop(ctx context.Context) {
	deadline, ok := ctx.Deadline()
	if ok {
		n.log.Info().Dur("timeout", time.Until(deadline)).Msg("Stopping managers with a grace period")
	} else {
		n.log.Info().Msg("Stopping managers with no deadline")
	}

	for _, m := range n.managers {
		m.Stop()
	}

	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		n.log.Info().Msg("Managers have stopped within the grace period")
		break
	case <-ctx.Done():
		n.log.Warn().Msg("Timeout while waiting, aborting the remaining managers")
		break
	}

	n.cancel()
	<-done
	n.log.Info().Msg("Stopped")
}
