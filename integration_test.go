//go:build integration

package gokogeri_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/kapvode/gokogeri"
	"github.com/kapvode/gokogeri/internal/redisutil"
	"github.com/kapvode/gokogeri/redis"
)

func TestDoMany(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cm := redis.NewConnManager(testConfig())
	defer cm.Close()

	t.Run("no error", func(t *testing.T) {
		assert := require.New(t)
		conn, err := cm.Conn(ctx)
		assert.NoError(err)
		defer conn.Close()

		err = conn.Send("LPUSH", "l1", "1")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l2", "2")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l3", "3")
		assert.NoError(err)

		_, err = redisutil.DoMany(conn, 3)
		assert.NoError(err)
	})

	t.Run("error in last command", func(t *testing.T) {
		assert := require.New(t)
		conn, err := cm.Conn(ctx)
		assert.NoError(err)
		defer conn.Close()

		err = conn.Send("LPUSH", "l1", "1")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l2", "2")
		assert.NoError(err)

		err = conn.Send("SADD", "l3", "3")
		assert.NoError(err)

		_, err = redisutil.DoMany(conn, 3)
		assert.Error(err)
		assert.ErrorContains(err, "reply 2:")
		assert.ErrorContains(err, "WRONGTYPE")
	})

	t.Run("error in first command", func(t *testing.T) {
		assert := require.New(t)
		conn, err := cm.Conn(ctx)
		assert.NoError(err)
		defer conn.Close()

		err = conn.Send("SADD", "l1", "1")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l2", "2")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l3", "3")
		assert.NoError(err)

		_, err = redisutil.DoMany(conn, 3)
		assert.Error(err)
		assert.ErrorContains(err, "reply 0:")
		assert.ErrorContains(err, "WRONGTYPE")
	})

	t.Run("unexpected number of replies", func(t *testing.T) {
		assert := require.New(t)
		conn, err := cm.Conn(ctx)
		assert.NoError(err)
		defer conn.Close()

		err = conn.Send("LPUSH", "l1", "1")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l2", "2")
		assert.NoError(err)

		err = conn.Send("LPUSH", "l3", "3")
		assert.NoError(err)

		_, err = redisutil.DoMany(conn, 2)
		assert.Error(err)
		assert.Equal("expect replies: 2, got 3", err.Error())
	})
}

func TestNoOpJobSuccess(t *testing.T) {
	// We enqueue a job that does nothing but finishes successfully. We confirm that the job the worker got matches
	// expectations.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	cm := redis.NewConnManager(testConfig())
	defer cm.Close()

	var workerJob *gokogeri.Job
	workerDone := make(chan struct{})

	node := gokogeri.NewNode(zerolog.Nop(), cm, 10)
	node.ProcessQueues(
		gokogeri.OrderedQueueSet{"default"},
		gokogeri.WorkerFunc(func(ctx context.Context, j *gokogeri.Job) error {
			defer close(workerDone)
			workerJob = j
			return nil
		}),
		1,
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		node.Run()
	}()

	now := time.Now()
	assert := require.New(t)

	job := gokogeri.Job{}
	job.SetClass("TestJob")

	enqueuer := gokogeri.NewEnqueuer(cm)
	err := enqueuer.Enqueue(ctx, &job)
	assert.NoError(err)

	select {
	case <-ctx.Done():
		assert.NoError(ctx.Err()) // fail on timeout
	case <-workerDone:
	}

	node.Stop(ctx)
	wg.Wait()
	assert.NoError(ctx.Err())

	assert.NotNil(workerJob)
	assert.Equal("default", workerJob.Queue())
	assert.Equal("TestJob", workerJob.Class())
	assert.Len(workerJob.ID(), 24)
	assert.Len(workerJob.Args(), 0)
	assert.WithinDuration(now, workerJob.CreatedAt(), time.Second)
	assert.Equal(workerJob.CreatedAt(), workerJob.EnqueuedAt())
}

func testConfig() *redis.Config {
	cfg := redis.NewDefaultConfig()
	cfg.URL = "redis://localhost/10"
	return cfg
}
