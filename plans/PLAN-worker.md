# Package Plan: pkg/worker

## Overview

A worker pool package for controlled concurrent task execution. Provides bounded concurrency, graceful shutdown, priority queues, and task lifecycle management for CPU-bound and I/O-bound workloads.

## Goals

1. **Bounded Concurrency** - Limit concurrent workers
2. **Priority Queues** - Process high-priority tasks first
3. **Graceful Shutdown** - Complete in-flight tasks
4. **Task Lifecycle** - Retry, timeout, cancellation
5. **Metrics** - Queue depth, processing time, throughput
6. **Rate Limiting** - Control task submission rate
7. **Zero Dependencies** - Pure Go implementation

## Architecture

```
pkg/worker/
├── worker.go             # Core Pool interface
├── config.go             # Configuration
├── options.go            # Functional options
├── task.go               # Task definition
├── queue.go              # Task queue
├── priority.go           # Priority queue
├── result.go             # Task result
├── group.go              # Wait group utilities
├── limiter.go            # Rate limiter
├── examples/
│   ├── basic/
│   ├── priority/
│   ├── batch/
│   └── graceful-shutdown/
└── README.md
```

## Core Interfaces

```go
package worker

import (
    "context"
    "time"
)

// Pool manages worker goroutines
type Pool interface {
    // Submit adds a task to the pool
    Submit(task Task) error

    // SubmitFunc adds a function as a task
    SubmitFunc(fn func(context.Context) error) error

    // SubmitWait submits and waits for result
    SubmitWait(ctx context.Context, task Task) (interface{}, error)

    // Start starts the worker pool
    Start(ctx context.Context) error

    // Stop gracefully stops the pool
    Stop(ctx context.Context) error

    // Resize changes the number of workers
    Resize(n int)

    // Stats returns pool statistics
    Stats() Stats

    // Wait waits for all tasks to complete
    Wait() error
}

// Task represents a unit of work
type Task interface {
    // Execute performs the task
    Execute(ctx context.Context) (interface{}, error)

    // ID returns the task identifier
    ID() string

    // Priority returns the task priority
    Priority() int
}

// TaskFunc is a function adapter for Task
type TaskFunc func(ctx context.Context) (interface{}, error)

// Result contains task execution result
type Result struct {
    TaskID    string
    Value     interface{}
    Error     error
    Duration  time.Duration
    StartedAt time.Time
}

// Stats contains pool statistics
type Stats struct {
    Workers       int
    ActiveWorkers int
    QueuedTasks   int
    CompletedTasks int64
    FailedTasks   int64
    TotalDuration time.Duration
}
```

## Configuration

```go
// Config holds worker pool configuration
type Config struct {
    // Number of workers
    Workers int `env:"WORKER_POOL_SIZE" default:"10"`

    // Queue capacity (0 = unbounded)
    QueueSize int `env:"WORKER_QUEUE_SIZE" default:"1000"`

    // Enable priority queue
    Priority bool `env:"WORKER_PRIORITY" default:"false"`

    // Task timeout (0 = no timeout)
    TaskTimeout time.Duration `env:"WORKER_TASK_TIMEOUT" default:"0"`

    // Graceful shutdown timeout
    ShutdownTimeout time.Duration `env:"WORKER_SHUTDOWN_TIMEOUT" default:"30s"`

    // Panic recovery
    PanicHandler func(interface{})
}

// Option configures the pool
type Option func(*poolConfig)

// WithWorkers sets the number of workers
func WithWorkers(n int) Option

// WithQueueSize sets the queue capacity
func WithQueueSize(size int) Option

// WithPriority enables priority queue
func WithPriority() Option

// WithTaskTimeout sets default task timeout
func WithTaskTimeout(timeout time.Duration) Option

// WithRetry configures task retry
func WithRetry(attempts int, delay time.Duration) Option

// WithRateLimit limits task processing rate
func WithRateLimit(rps float64) Option

// WithMetrics enables metrics collection
func WithMetrics(collector MetricsCollector) Option

// WithPanicHandler sets panic recovery handler
func WithPanicHandler(fn func(interface{})) Option
```

## Task Options

```go
// TaskOption configures individual tasks
type TaskOption func(*taskConfig)

// WithTaskID sets the task ID
func WithTaskID(id string) TaskOption

// WithTaskPriority sets the task priority
func WithTaskPriority(priority int) TaskOption

// WithTaskTimeout sets task-specific timeout
func WithTaskTimeout(timeout time.Duration) TaskOption

// WithTaskRetry sets task-specific retry
func WithTaskRetry(attempts int) TaskOption

// WithTaskCallback sets completion callback
func WithTaskCallback(fn func(Result)) TaskOption
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/user/core-backend/pkg/worker"
)

func main() {
    // Create pool with 5 workers
    pool := worker.New(worker.Config{
        Workers:   5,
        QueueSize: 100,
    })

    ctx := context.Background()
    pool.Start(ctx)
    defer pool.Stop(ctx)

    // Submit tasks
    for i := 0; i < 100; i++ {
        i := i
        pool.SubmitFunc(func(ctx context.Context) error {
            fmt.Printf("Processing task %d\n", i)
            time.Sleep(100 * time.Millisecond)
            return nil
        })
    }

    // Wait for completion
    pool.Wait()
}
```

### With Results

```go
func main() {
    pool := worker.New(cfg)
    pool.Start(ctx)
    defer pool.Stop(ctx)

    // Submit and wait for result
    result, err := pool.SubmitWait(ctx, worker.TaskFunc(
        func(ctx context.Context) (interface{}, error) {
            return computeExpensiveValue(), nil
        },
    ))

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Result:", result)
}
```

### Priority Queue

```go
func main() {
    pool := worker.New(worker.Config{
        Workers:  5,
        Priority: true,
    })
    pool.Start(ctx)

    // High priority (processed first)
    pool.Submit(worker.NewTask(
        func(ctx context.Context) (interface{}, error) {
            return processUrgent(), nil
        },
        worker.WithTaskPriority(10),
    ))

    // Normal priority
    pool.Submit(worker.NewTask(
        func(ctx context.Context) (interface{}, error) {
            return processNormal(), nil
        },
        worker.WithTaskPriority(5),
    ))

    // Low priority (processed last)
    pool.Submit(worker.NewTask(
        func(ctx context.Context) (interface{}, error) {
            return processBackground(), nil
        },
        worker.WithTaskPriority(1),
    ))
}
```

### Batch Processing

```go
func main() {
    pool := worker.New(cfg)
    pool.Start(ctx)

    // Create result channel
    results := make(chan worker.Result, len(items))

    // Submit batch
    for _, item := range items {
        item := item
        pool.Submit(worker.NewTask(
            func(ctx context.Context) (interface{}, error) {
                return processItem(item), nil
            },
            worker.WithTaskCallback(func(r worker.Result) {
                results <- r
            }),
        ))
    }

    // Collect results
    for i := 0; i < len(items); i++ {
        result := <-results
        if result.Error != nil {
            log.Printf("Task %s failed: %v", result.TaskID, result.Error)
        }
    }
}
```

### With Retry

```go
func main() {
    pool := worker.New(cfg,
        worker.WithRetry(3, time.Second), // Retry 3 times with 1s delay
    )
    pool.Start(ctx)

    pool.Submit(worker.NewTask(
        func(ctx context.Context) (interface{}, error) {
            return callExternalAPI() // May fail transiently
        },
        worker.WithTaskRetry(5), // Override: 5 retries for this task
    ))
}
```

### Rate Limited

```go
func main() {
    pool := worker.New(cfg,
        worker.WithRateLimit(10), // Max 10 tasks per second
    )
    pool.Start(ctx)

    // Tasks are processed at controlled rate
    for i := 0; i < 1000; i++ {
        pool.SubmitFunc(func(ctx context.Context) error {
            return callRateLimitedAPI()
        })
    }
}
```

### Graceful Shutdown

```go
func main() {
    pool := worker.New(worker.Config{
        Workers:         10,
        ShutdownTimeout: 30 * time.Second,
    })

    ctx := context.Background()
    pool.Start(ctx)

    // Submit many tasks
    for i := 0; i < 1000; i++ {
        pool.SubmitFunc(processTask)
    }

    // Handle shutdown signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    <-sigCh

    // Graceful shutdown
    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    if err := pool.Stop(shutdownCtx); err != nil {
        log.Printf("Shutdown timeout: %v", err)
    }
}
```

### Dynamic Resizing

```go
func main() {
    pool := worker.New(worker.Config{
        Workers: 5,
    })
    pool.Start(ctx)

    // Monitor and resize based on load
    go func() {
        for {
            stats := pool.Stats()

            // Scale up if queue is backing up
            if stats.QueuedTasks > 100 && stats.Workers < 20 {
                pool.Resize(stats.Workers + 5)
            }

            // Scale down if idle
            if stats.QueuedTasks == 0 && stats.ActiveWorkers == 0 {
                pool.Resize(5) // Minimum workers
            }

            time.Sleep(10 * time.Second)
        }
    }()
}
```

### Error Group Pattern

```go
import "github.com/user/core-backend/pkg/worker"

func main() {
    pool := worker.New(cfg)
    pool.Start(ctx)

    // Create error group
    g := worker.NewGroup(pool)

    // Submit tasks
    g.Go(func(ctx context.Context) error {
        return fetchDataA(ctx)
    })

    g.Go(func(ctx context.Context) error {
        return fetchDataB(ctx)
    })

    g.Go(func(ctx context.Context) error {
        return fetchDataC(ctx)
    })

    // Wait for all, return first error
    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }
}
```

## Dependencies

- **Required:** None (pure Go)

## Implementation Phases

### Phase 1: Core Pool
1. Basic worker pool
2. Task submission
3. Bounded queue

### Phase 2: Task Management
1. Task interface
2. Result handling
3. Callbacks

### Phase 3: Priority Queue
1. Priority queue implementation
2. Priority-based scheduling

### Phase 4: Lifecycle
1. Graceful shutdown
2. Task timeout
3. Retry logic

### Phase 5: Advanced Features
1. Rate limiting
2. Dynamic resizing
3. Metrics collection

### Phase 6: Documentation
1. README
2. Examples
3. Best practices
