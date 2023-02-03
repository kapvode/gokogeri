package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/kapvode/gokogeri"
	"github.com/kapvode/gokogeri/redis"
)

func devLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	return zerolog.New(output).Level(zerolog.InfoLevel).With().Timestamp().Logger()
}

func main() {
	cfg := redis.NewDefaultConfig()
	cfg.URL = "redis://localhost/4"

	cm := redis.NewConnManager(cfg)
	defer cm.Close()

	ctx := context.Background()

	var job1 gokogeri.Job
	job1.SetQueue("critical")
	job1.SetClass("CriticalJob")

	var job2 gokogeri.Job
	job2.SetQueue("low_priority")
	job2.SetClass("LowPriorityJob")

	enqueuer := gokogeri.NewEnqueuer(cm)

	err := enqueuer.Enqueue(ctx, &job1)
	if err != nil {
		fmt.Println("enqueue job1:", err)
	}
	err = enqueuer.Enqueue(ctx, &job2)
	if err != nil {
		fmt.Println("enqueue job2:", err)
	}

	logger := gokogeri.DefaultLogger()
	// Alternatively
	// logger := zerolog.Nop()
	// logger := devLogger()

	node := gokogeri.NewNode(logger, cm, cfg.LongPollTimeout)

	node.ProcessQueues(
		gokogeri.OrderedQueueSet{"critical"},
		gokogeri.WorkerFunc(func(ctx context.Context, j *gokogeri.Job) error {
			fmt.Printf("processing job: queue=%s class=%s id=%s\n", j.Queue(), j.Class(), j.ID())

			time.Sleep(time.Second * 5)

			fmt.Printf("job done: queue=%s class=%s id=%s\n", j.Queue(), j.Class(), j.ID())

			return nil
		}),
		1,
	)

	qs := gokogeri.NewRandomQueueSet()
	qs.Add("low_priority", 1)
	qs.Add("high_priority", 3)

	node.ProcessQueues(
		qs,
		gokogeri.WorkerFunc(func(ctx context.Context, j *gokogeri.Job) error {
			fmt.Printf("processing job: queue=%s class=%s id=%s\n", j.Queue(), j.Class(), j.ID())

			time.Sleep(time.Second * 7)

			fmt.Printf("job done: queue=%s class=%s id=%s\n", j.Queue(), j.Class(), j.ID())

			return nil
		}),
		5,
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		node.Run()
	}()

	// Give the jobs a chance to start running before we initiate shutdown, since this is just an example.
	time.Sleep(time.Second)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	node.Stop(shutdownCtx)

	wg.Wait()
}
