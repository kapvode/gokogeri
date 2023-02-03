# gokogeri

gokogeri is a Go package for asynchronous job processing that tries to implement some of the most useful parts of [Sidekiq](https://github.com/mperham/sidekiq) in a compatible way.

It is still in the early stages of development and is not ready to be used.

## Quick Start

You can find a complete example in [example/example.go](example/example.go).

### Configuration

Start with the default configuration and adjust it according to your needs.

```go
cfg := redis.NewDefaultConfig()
cfg.URL = "redis://localhost/4"
```

### Logging

Use the default logger

```go
logger := gokogeri.DefaultLogger()
```

or disable logging

```go
logger := zerolog.Nop()
```

### Connecting to Redis

```go
cm := redis.NewConnManager(cfg)
defer cm.Close()
```

### Adding jobs

```go
var job gokogeri.Job
job.SetQueue("critical")
job.SetClass("CriticalJob")

enqueuer := gokogeri.NewEnqueuer(cm)
enqueuer.Enqueue(ctx, &job)
```

### Processing jobs

Create a node, which represents an instance of a server that is processing jobs.

```go
node := gokogeri.NewNode(logger, cm, cfg.LongPollTimeout)
```

Define the queues that you want to process, along with the workers and the number of instances (goroutines) of those workers.

You should probably implement the `Worker` interface. The examples use `WorkerFunc`.

```go
// 1 dedicated goroutine for processing critical jobs

instances := 1
node.ProcessQueues(
    gokogeri.OrderedQueueSet{"critical"},
    gokogeri.WorkerFunc(func(ctx context.Context, j *gokogeri.Job) error {
        var err error

        // do the work

        return err
    }),
    instances,
)
```

Use an `OrderedQueueSet` to process queues in a strictly defined order of priorities.

Use a `RandomQueueSet` to process queues in random order, with the likelihood of each queue being checked first based on their relative weights.

```go
// 5 instances processing jobs from two queues, with weight ratios 3 to 1.
// The high_priority queue has a 75% chance of being checked first.
// The low_priority queue has a 25% chance of being checked first.

qs := gokogeri.NewRandomQueueSet()
qs.Add("low_priority", 1)
qs.Add("high_priority", 3)

instances := 5
node.ProcessQueues(
    qs,
    gokogeri.WorkerFunc(func(ctx context.Context, j *gokogeri.Job) error {
        var err error

        // do the work

        return err
    }),
    instances,
)
```

Run the node. This will block until the node is stopped, so you should probably run it in another gorutine.

```go
node.Run()
```

When you want to stop the node, call Stop, and pass a shutdown context as a timeout.

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
defer cancel()

node.Stop(shutdownCtx)
```

New jobs will not be taken from queues any more. Workers that are currently processing jobs will be given a grace period and allowed to finish. Once the Context provided to Stop expires, the Context passed to every Worker will be cancelled.
