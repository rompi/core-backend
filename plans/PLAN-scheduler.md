# Package Plan: pkg/scheduler

## Overview

A background job processing and scheduling package supporting cron-like scheduling, one-time jobs, and distributed job execution. Provides reliable job processing with retry, dead letter handling, and multiple storage backends.

## Goals

1. **Cron Scheduling** - Schedule recurring jobs with cron syntax
2. **One-Time Jobs** - Schedule jobs to run once at a specific time
3. **Job Queuing** - Queue jobs for immediate or delayed processing
4. **Distributed Execution** - Leader election for multi-instance deployments
5. **Persistence** - Store jobs in PostgreSQL or Redis
6. **Reliability** - Retry with backoff, dead letter queue
7. **Observability** - Hooks for logging, metrics, and tracing

## Architecture

```
pkg/scheduler/
├── scheduler.go          # Core Scheduler interface
├── config.go             # Configuration with env support
├── options.go            # Functional options
├── job.go                # Job type definition
├── cron.go               # Cron expression parser
├── worker.go             # Worker pool for job execution
├── errors.go             # Custom error types
├── store/
│   ├── store.go          # Storage interface
│   ├── memory.go         # In-memory storage (testing/dev)
│   ├── postgres.go       # PostgreSQL storage
│   ├── redis.go          # Redis storage
│   └── store_test.go
├── distributed/
│   ├── leader.go         # Leader election interface
│   ├── postgres.go       # PostgreSQL-based leader election
│   └── redis.go          # Redis-based leader election
├── middleware/
│   ├── logging.go        # Logging middleware
│   ├── tracing.go        # Tracing middleware
│   ├── metrics.go        # Metrics middleware
│   └── recovery.go       # Panic recovery
├── examples/
│   ├── basic/
│   ├── cron/
│   ├── distributed/
│   └── with-postgres/
└── README.md
```

## Core Interfaces

```go
package scheduler

import (
    "context"
    "time"
)

// Scheduler manages job scheduling and execution
type Scheduler interface {
    // Schedule adds a recurring job with cron schedule
    Schedule(name string, schedule string, handler Handler, opts ...JobOption) error

    // ScheduleFunc adds a recurring job with function
    ScheduleFunc(name string, schedule string, fn func(context.Context) error, opts ...JobOption) error

    // Every schedules a job at fixed intervals
    Every(interval time.Duration) *JobBuilder

    // At schedules a one-time job
    At(when time.Time) *JobBuilder

    // Enqueue adds a job to the queue for immediate processing
    Enqueue(ctx context.Context, job *Job) error

    // EnqueueAt adds a job to the queue for delayed processing
    EnqueueAt(ctx context.Context, job *Job, when time.Time) error

    // Start begins processing jobs
    Start(ctx context.Context) error

    // Stop gracefully stops the scheduler
    Stop(ctx context.Context) error

    // Running returns true if scheduler is running
    Running() bool
}

// Handler processes jobs
type Handler interface {
    Handle(ctx context.Context, job *Job) error
}

// HandlerFunc is a function adapter for Handler
type HandlerFunc func(ctx context.Context, job *Job) error

func (f HandlerFunc) Handle(ctx context.Context, job *Job) error {
    return f(ctx, job)
}

// Job represents a scheduled or queued job
type Job struct {
    // ID is unique identifier
    ID string

    // Name identifies the job type
    Name string

    // Payload is job-specific data (JSON)
    Payload []byte

    // Queue is the queue name
    Queue string

    // Priority (higher = more important)
    Priority int

    // MaxRetries is maximum retry attempts
    MaxRetries int

    // RetryCount is current retry count
    RetryCount int

    // Timeout for job execution
    Timeout time.Duration

    // ScheduledAt is when job should run
    ScheduledAt time.Time

    // CreatedAt is when job was created
    CreatedAt time.Time

    // StartedAt is when job started executing
    StartedAt *time.Time

    // CompletedAt is when job completed
    CompletedAt *time.Time

    // Error is the last error message
    Error string

    // Status is job status
    Status JobStatus

    // Metadata for custom data
    Metadata map[string]string
}

// JobStatus represents job state
type JobStatus string

const (
    JobStatusPending   JobStatus = "pending"
    JobStatusScheduled JobStatus = "scheduled"
    JobStatusRunning   JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed    JobStatus = "failed"
    JobStatusDead      JobStatus = "dead" // Max retries exceeded
)
```

## Configuration

```go
// Config holds scheduler configuration
type Config struct {
    // Store type: "memory", "postgres", "redis"
    Store string `env:"SCHEDULER_STORE" default:"memory"`

    // Worker pool size
    Workers int `env:"SCHEDULER_WORKERS" default:"5"`

    // Poll interval for checking new jobs
    PollInterval time.Duration `env:"SCHEDULER_POLL_INTERVAL" default:"1s"`

    // Default job timeout
    DefaultTimeout time.Duration `env:"SCHEDULER_DEFAULT_TIMEOUT" default:"5m"`

    // Default retry settings
    Retry RetryConfig

    // Dead letter queue settings
    DeadLetter DeadLetterConfig

    // Distributed mode settings
    Distributed DistributedConfig
}

type RetryConfig struct {
    // Maximum retry attempts
    MaxAttempts int `env:"SCHEDULER_RETRY_MAX_ATTEMPTS" default:"3"`

    // Initial delay between retries
    InitialDelay time.Duration `env:"SCHEDULER_RETRY_INITIAL_DELAY" default:"1s"`

    // Maximum delay between retries
    MaxDelay time.Duration `env:"SCHEDULER_RETRY_MAX_DELAY" default:"1h"`

    // Backoff multiplier
    Multiplier float64 `env:"SCHEDULER_RETRY_MULTIPLIER" default:"2.0"`
}

type DeadLetterConfig struct {
    // Enable dead letter queue
    Enabled bool `env:"SCHEDULER_DLQ_ENABLED" default:"true"`

    // Queue name for dead jobs
    Queue string `env:"SCHEDULER_DLQ_QUEUE" default:"dead"`

    // Retention period for dead jobs
    Retention time.Duration `env:"SCHEDULER_DLQ_RETENTION" default:"168h"` // 7 days
}

type DistributedConfig struct {
    // Enable distributed mode (leader election)
    Enabled bool `env:"SCHEDULER_DISTRIBUTED" default:"false"`

    // Instance ID (auto-generated if empty)
    InstanceID string `env:"SCHEDULER_INSTANCE_ID" default:""`

    // Leader lock TTL
    LockTTL time.Duration `env:"SCHEDULER_LOCK_TTL" default:"30s"`

    // Leader lock refresh interval
    LockRefresh time.Duration `env:"SCHEDULER_LOCK_REFRESH" default:"10s"`
}
```

## Storage Configurations

### PostgreSQL

```go
type PostgresConfig struct {
    // Connection string or use postgres package
    ConnectionString string `env:"SCHEDULER_POSTGRES_URL" default:""`

    // Table name prefix
    TablePrefix string `env:"SCHEDULER_POSTGRES_TABLE_PREFIX" default:"scheduler_"`

    // Schema name
    Schema string `env:"SCHEDULER_POSTGRES_SCHEMA" default:"public"`
}

// Required tables:
// - scheduler_jobs: job storage
// - scheduler_schedules: cron schedules
// - scheduler_locks: distributed locks
```

### Redis

```go
type RedisConfig struct {
    // Redis URL
    URL string `env:"SCHEDULER_REDIS_URL" default:"redis://localhost:6379"`

    // Key prefix
    KeyPrefix string `env:"SCHEDULER_REDIS_KEY_PREFIX" default:"scheduler:"`

    // Database number
    DB int `env:"SCHEDULER_REDIS_DB" default:"0"`
}
```

## Job Builder

```go
// JobBuilder provides fluent job construction
type JobBuilder struct {
    scheduler *Scheduler
    interval  time.Duration
    at        time.Time
}

// Do sets the job handler
func (b *JobBuilder) Do(name string, handler Handler, opts ...JobOption) error

// DoFunc sets the job function
func (b *JobBuilder) DoFunc(name string, fn func(context.Context) error, opts ...JobOption) error

// Every creates interval-based scheduling
func (s *Scheduler) Every(interval time.Duration) *JobBuilder

// Fluent helpers
func (b *JobBuilder) Second() *JobBuilder
func (b *JobBuilder) Seconds() *JobBuilder
func (b *JobBuilder) Minute() *JobBuilder
func (b *JobBuilder) Minutes() *JobBuilder
func (b *JobBuilder) Hour() *JobBuilder
func (b *JobBuilder) Hours() *JobBuilder
func (b *JobBuilder) Day() *JobBuilder
func (b *JobBuilder) Days() *JobBuilder
func (b *JobBuilder) Week() *JobBuilder
func (b *JobBuilder) Weeks() *JobBuilder

// At specific time
func (b *JobBuilder) At(time string) *JobBuilder // "14:30"
func (b *JobBuilder) Monday() *JobBuilder
func (b *JobBuilder) Tuesday() *JobBuilder
// ... other days
```

## Job Options

```go
// JobOption configures a job
type JobOption func(*jobOptions)

// WithQueue sets the queue name
func WithQueue(queue string) JobOption

// WithPriority sets job priority
func WithPriority(priority int) JobOption

// WithTimeout sets execution timeout
func WithTimeout(timeout time.Duration) JobOption

// WithMaxRetries sets maximum retry attempts
func WithMaxRetries(retries int) JobOption

// WithPayload sets job payload
func WithPayload(payload interface{}) JobOption

// WithMetadata sets job metadata
func WithMetadata(meta map[string]string) JobOption

// WithDelay sets initial delay before first run
func WithDelay(delay time.Duration) JobOption

// WithUniqueKey prevents duplicate jobs
func WithUniqueKey(key string) JobOption
```

## Cron Expression Support

```go
// Cron expression format:
// ┌───────────── second (0 - 59) [optional]
// │ ┌───────────── minute (0 - 59)
// │ │ ┌───────────── hour (0 - 23)
// │ │ │ ┌───────────── day of month (1 - 31)
// │ │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ │ ┌───────────── day of week (0 - 6) (Sunday = 0)
// │ │ │ │ │ │
// * * * * * *

// Predefined schedules
const (
    EveryMinute   = "* * * * *"
    Hourly        = "0 * * * *"
    Daily         = "0 0 * * *"
    Weekly        = "0 0 * * 0"
    Monthly       = "0 0 1 * *"
    Yearly        = "0 0 1 1 *"
)

// ParseCron parses a cron expression
func ParseCron(expr string) (Schedule, error)

// Schedule represents a cron schedule
type Schedule interface {
    Next(time.Time) time.Time
}
```

## Storage Interface

```go
// Store persists jobs and schedules
type Store interface {
    // Job operations
    CreateJob(ctx context.Context, job *Job) error
    GetJob(ctx context.Context, id string) (*Job, error)
    UpdateJob(ctx context.Context, job *Job) error
    DeleteJob(ctx context.Context, id string) error

    // Queue operations
    FetchJobs(ctx context.Context, queue string, limit int) ([]*Job, error)
    AckJob(ctx context.Context, id string) error
    NackJob(ctx context.Context, id string, err error) error

    // Schedule operations
    SaveSchedule(ctx context.Context, schedule *ScheduledJob) error
    GetSchedules(ctx context.Context) ([]*ScheduledJob, error)
    DeleteSchedule(ctx context.Context, name string) error

    // Maintenance
    CleanupCompleted(ctx context.Context, olderThan time.Duration) error
    Stats(ctx context.Context) (*Stats, error)

    // Close releases resources
    Close() error
}

// Stats provides queue statistics
type Stats struct {
    Pending   int64
    Running   int64
    Completed int64
    Failed    int64
    Dead      int64
    Scheduled int64
}
```

## Leader Election

```go
// Leader manages distributed scheduling
type Leader interface {
    // IsLeader returns true if this instance is the leader
    IsLeader() bool

    // Acquire attempts to acquire leadership
    Acquire(ctx context.Context) error

    // Release releases leadership
    Release(ctx context.Context) error

    // OnLeaderChange registers callback for leadership changes
    OnLeaderChange(callback func(isLeader bool))
}
```

## Middleware

```go
// Middleware wraps job execution
type Middleware func(Handler) Handler

// Chain combines middleware
func Chain(middlewares ...Middleware) Middleware

// Logging logs job execution
func Logging(logger Logger) Middleware

// Tracing adds distributed tracing
func Tracing(tracer Tracer) Middleware

// Metrics records job metrics
func Metrics(collector MetricsCollector) Middleware

// Recovery recovers from panics
func Recovery(logger Logger) Middleware

// Timeout enforces execution timeout
func Timeout(timeout time.Duration) Middleware
```

## Error Handling

```go
var (
    // ErrJobNotFound is returned when job doesn't exist
    ErrJobNotFound = errors.New("scheduler: job not found")

    // ErrScheduleInvalid is returned for invalid cron expressions
    ErrScheduleInvalid = errors.New("scheduler: invalid schedule")

    // ErrJobAlreadyExists is returned for duplicate unique jobs
    ErrJobAlreadyExists = errors.New("scheduler: job already exists")

    // ErrNotLeader is returned when not leader in distributed mode
    ErrNotLeader = errors.New("scheduler: not leader")

    // ErrShuttingDown is returned during shutdown
    ErrShuttingDown = errors.New("scheduler: shutting down")
)

// RetryableError indicates job should be retried
type RetryableError struct {
    Err   error
    Delay time.Duration // Custom delay, 0 = use default
}

func (e RetryableError) Error() string { return e.Err.Error() }
func (e RetryableError) Unwrap() error { return e.Err }

// Retry returns a retryable error
func Retry(err error) error { return RetryableError{Err: err} }

// RetryAfter returns a retryable error with custom delay
func RetryAfter(err error, delay time.Duration) error
```

## Usage Examples

### Basic Scheduling

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/scheduler"
)

func main() {
    // Create scheduler
    s, err := scheduler.New(scheduler.Config{
        Store:   "memory",
        Workers: 5,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Schedule a cron job
    err = s.Schedule("cleanup", "0 0 * * *", // Daily at midnight
        scheduler.HandlerFunc(func(ctx context.Context, job *scheduler.Job) error {
            log.Println("Running cleanup...")
            return nil
        }),
    )

    // Schedule with fluent API
    s.Every(5).Minutes().DoFunc("healthcheck", func(ctx context.Context) error {
        log.Println("Running health check...")
        return nil
    })

    // Start scheduler
    ctx := context.Background()
    if err := s.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Wait for interrupt
    <-ctx.Done()
    s.Stop(context.Background())
}
```

### Queueing Jobs

```go
func main() {
    s, _ := scheduler.New(cfg)
    s.Start(ctx)

    // Register handler for job type
    s.Register("send_email", scheduler.HandlerFunc(
        func(ctx context.Context, job *scheduler.Job) error {
            var payload EmailPayload
            json.Unmarshal(job.Payload, &payload)
            return sendEmail(payload)
        },
    ))

    // Enqueue job for immediate processing
    job := &scheduler.Job{
        Name: "send_email",
        Payload: mustMarshal(EmailPayload{
            To:      "user@example.com",
            Subject: "Welcome!",
        }),
    }
    s.Enqueue(ctx, job)

    // Enqueue job for later
    s.EnqueueAt(ctx, job, time.Now().Add(time.Hour))
}
```

### With PostgreSQL Store

```go
import (
    "github.com/user/core-backend/pkg/scheduler"
    "github.com/user/core-backend/pkg/scheduler/store"
)

func main() {
    // Create PostgreSQL store
    st, err := store.NewPostgres(store.PostgresConfig{
        ConnectionString: "postgres://user:pass@localhost/db",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create scheduler with store
    s, err := scheduler.New(scheduler.Config{
        Store: "postgres",
    }, scheduler.WithStore(st))

    s.Start(ctx)
}
```

### Distributed Mode

```go
func main() {
    s, err := scheduler.New(scheduler.Config{
        Store:   "redis",
        Workers: 10,
        Distributed: scheduler.DistributedConfig{
            Enabled:    true,
            InstanceID: os.Getenv("HOSTNAME"),
            LockTTL:    30 * time.Second,
        },
    })

    // Only leader runs scheduled jobs
    // All instances process queue jobs
    s.Schedule("daily_report", "0 9 * * *", generateReport)

    s.Start(ctx)
}
```

### With Retry and Dead Letter

```go
func main() {
    s, _ := scheduler.New(scheduler.Config{
        Retry: scheduler.RetryConfig{
            MaxAttempts:  5,
            InitialDelay: time.Second,
            MaxDelay:     time.Hour,
            Multiplier:   2.0,
        },
        DeadLetter: scheduler.DeadLetterConfig{
            Enabled: true,
            Queue:   "dead_jobs",
        },
    })

    s.Register("process_order", scheduler.HandlerFunc(
        func(ctx context.Context, job *scheduler.Job) error {
            err := processOrder(job)
            if err != nil {
                // Return retryable error
                return scheduler.Retry(err)
            }
            return nil
        },
    ), scheduler.WithMaxRetries(3))

    // Handle dead jobs
    s.Register("dead_jobs", scheduler.HandlerFunc(
        func(ctx context.Context, job *scheduler.Job) error {
            log.Printf("Dead job: %s, error: %s", job.ID, job.Error)
            alertOps(job)
            return nil
        },
    ))
}
```

### With Middleware

```go
func main() {
    s, _ := scheduler.New(cfg)

    // Add global middleware
    s.Use(
        scheduler.Recovery(logger),
        scheduler.Logging(logger),
        scheduler.Tracing(tracer),
        scheduler.Metrics(prometheus),
    )

    s.Schedule("cleanup", "0 * * * *", cleanupHandler)
    s.Start(ctx)
}
```

### One-Time Jobs

```go
func main() {
    s, _ := scheduler.New(cfg)
    s.Start(ctx)

    // Schedule one-time job
    s.At(time.Now().Add(24 * time.Hour)).DoFunc("birthday_email",
        func(ctx context.Context) error {
            return sendBirthdayEmail(userID)
        },
    )

    // Or with fluent API
    s.Every(1).Day().At("09:00").DoFunc("morning_report", generateMorningReport)
}
```

### Typed Jobs with Generics

```go
// TypedHandler provides type-safe job handling
type TypedHandler[T any] struct {
    handler func(context.Context, T) error
}

func NewTypedHandler[T any](fn func(context.Context, T) error) *TypedHandler[T] {
    return &TypedHandler[T]{handler: fn}
}

func (h *TypedHandler[T]) Handle(ctx context.Context, job *Job) error {
    var payload T
    if err := json.Unmarshal(job.Payload, &payload); err != nil {
        return err
    }
    return h.handler(ctx, payload)
}

// Usage
type OrderPayload struct {
    OrderID string
    UserID  string
}

s.Register("process_order", NewTypedHandler(
    func(ctx context.Context, order OrderPayload) error {
        return processOrder(order)
    },
))
```

## Health Check

```go
// HealthCheck returns a health check function
func (s *Scheduler) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        if !s.Running() {
            return errors.New("scheduler not running")
        }
        return s.store.Ping(ctx)
    }
}
```

## Observability Hooks

```go
// Hook interface for observability
type Hook interface {
    BeforeJob(ctx context.Context, job *Job)
    AfterJob(ctx context.Context, job *Job, err error, duration time.Duration)
    OnEnqueue(ctx context.Context, job *Job)
    OnRetry(ctx context.Context, job *Job, attempt int, err error)
    OnDead(ctx context.Context, job *Job, err error)
}

// WithHook adds observability hooks
func WithHook(hook Hook) Option
```

## Dependencies

- **Required:** None (in-memory implementation)
- **Optional:**
  - `github.com/robfig/cron/v3` for cron parsing (or custom implementation)
  - `github.com/jackc/pgx/v5` for PostgreSQL store
  - `github.com/redis/go-redis/v9` for Redis store

## Test Coverage Requirements

- Unit tests for all public functions
- Cron parser tests with edge cases
- Integration tests with PostgreSQL and Redis
- Concurrency tests
- Leader election tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Memory Store
1. Define Scheduler, Job, Handler interfaces
2. Implement in-memory store
3. Worker pool implementation
4. Basic scheduling

### Phase 2: Cron Support
1. Cron expression parser
2. Schedule interface
3. Fluent job builder
4. Time zone support

### Phase 3: PostgreSQL Store
1. Implement PostgreSQL store
2. Database schema
3. Job locking for concurrent access
4. Integration tests

### Phase 4: Redis Store
1. Implement Redis store
2. Atomic operations with Lua scripts
3. Integration tests

### Phase 5: Distributed Mode
1. Leader election interface
2. PostgreSQL-based election
3. Redis-based election
4. Multi-instance tests

### Phase 6: Advanced Features
1. Middleware system
2. Retry with exponential backoff
3. Dead letter queue
4. Job uniqueness

### Phase 7: Documentation & Examples
1. README with full documentation
2. Basic example
3. Cron example
4. Distributed example
