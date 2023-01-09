package gokogeri

import (
	"context"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
)

type dequeuerFactory struct {
	log        zerolog.Logger
	cp         ConnProvider
	popTimeout int // seconds
}

func newDequeuerFactory(log zerolog.Logger, cp ConnProvider, popTimeout int) *dequeuerFactory {
	return &dequeuerFactory{
		log:        log,
		cp:         cp,
		popTimeout: popTimeout,
	}
}

func (f *dequeuerFactory) newDequeuer(qset QueueSet) *dequeuer {
	dq := &dequeuer{
		cp:         f.cp,
		qset:       qset,
		popTimeout: f.popTimeout,
	}
	dq.ctx, dq.cancel = context.WithCancel(context.Background())
	dq.C = make(chan workItem)
	dq.log = f.log.With().Str("component", "dequeuer").Strs("queue_set", qset.Names()).Logger()
	return dq
}

type dequeuer struct {
	log zerolog.Logger
	cp  ConnProvider

	ctx    context.Context
	cancel context.CancelFunc

	qset       QueueSet
	popArgs    []interface{}
	popTimeout int // seconds

	// mu guards changes to the value of conn, not its use. Concurrent access is expected only when closing, between
	// conn.Close and conn.Do.
	mu   sync.Mutex
	conn redis.Conn

	// C is the channel on which the dequeuer will send what it reads from one of the queues in the set. It will close
	// the channel when it is shut down, indicating there will be no more work.
	C chan workItem
}

// Run blocks until the dequeuer is stopped.
func (dq *dequeuer) Run() {
	defer close(dq.C)

	// - Connect in a loop.
	// - Pause between attempts if there is a problem.
	// - When we have a connection, run the poll / pop loop.
	// - Send what we popped on the channel.
	// - When asked to stop, close the channel.
	dq.connectLoop()

	dq.log.Debug().Msg("Stopped")
}

// Stop initiates shutdown, but does not block.
func (dq *dequeuer) Stop() {
	dq.log.Debug().Msg("Stopping")
	dq.cancel()
	dq.closeConn()
}

func (dq *dequeuer) connectLoop() {
	for {
		select {
		case <-dq.ctx.Done():
			return
		default:
			dq.log.Info().Msg("Connecting")
			if err := dq.connect(); err != nil {
				if dq.notClosing() {
					time.Sleep(time.Second)
				}
				continue
			}
		}

		dq.log.Info().Msg("Connected")
		dq.readLoop()
	}
}

func (dq *dequeuer) connect() error {
	conn, err := dq.cp.DialLongPoll(dq.ctx)
	if err != nil {
		dq.log.Error().Err(err).Msg("Failed to connect to Redis")
		return err
	}

	dq.setConn(conn)
	return nil
}

func (dq *dequeuer) readLoop() {
	defer func() {
		dq.conn.Close()
		dq.setConn(nil)
	}()

	for {
		select {
		case <-dq.ctx.Done():
			return
		default:
			dq.log.Trace().Msg("BRPOP")
			results, err := redis.ByteSlices(dq.conn.Do("BRPOP", dq.getPopArgs()...))
			if err != nil {
				if err == redis.ErrNil {
					dq.log.Trace().Msg("BRPOP timeout")
					continue
				}
				if dq.notClosing() {
					dq.log.Error().Err(err).Msg("Failed to read from the queue set")
				}
				return
			}
			if len(results) != 2 {
				dq.log.Error().Msgf("Expected 2 results, got %d", len(results))
				return
			}
			dq.C <- workItem{
				Q: string(results[0]),
				P: results[1],
			}
		}
	}
}

func (dq *dequeuer) setConn(conn redis.Conn) {
	dq.mu.Lock()
	dq.conn = conn
	dq.mu.Unlock()
}

func (dq *dequeuer) closeConn() {
	dq.mu.Lock()
	if dq.conn != nil {
		dq.conn.Close()
	}
	dq.mu.Unlock()
}

func (dq *dequeuer) notClosing() bool {
	return dq.ctx.Err() == nil
}

func (dq *dequeuer) getPopArgs() []interface{} {
	queues := dq.qset.GetQueues()
	if dq.popArgs == nil {
		// size = number of queues + timeout
		dq.popArgs = make([]interface{}, 0, len(queues)+1)
	}
	dq.popArgs = dq.popArgs[:0]
	for _, q := range queues {
		dq.popArgs = append(dq.popArgs, "queue:"+q)
	}
	dq.popArgs = append(dq.popArgs, dq.popTimeout)
	return dq.popArgs
}
