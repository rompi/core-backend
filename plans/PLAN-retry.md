# Package Plan: pkg/retry

## Overview

A retry utility package providing configurable retry strategies with backoff algorithms. Supports exponential backoff, jitter, circuit breaking integration, and context-aware cancellation for building resilient applications.

## Goals

1. **Multiple Strategies** - Exponential, linear, constant backoff
2. **Jitter Support** - Prevent thundering herd
3. **Configurable Conditions** - Retry on specific errors
4. **Context Aware** - Cancellation and timeout support
5. **Hooks** - Before/after retry callbacks
6. **Zero Dependencies** - Pure Go implementation
7. **Generic Support** - Works with any operation

## Architecture

```
pkg/retry/
├── retry.go              # Core retry functions
├── config.go             # Configuration
├── backoff.go            # Backoff strategies
├── jitter.go             # Jitter algorithms
├── errors.go             # Error handling
├── hooks.go              # Retry hooks
├── examples/
│   ├── basic/
│   ├── http-client/
│   └── database/
└── README.md
```

## Core Interfaces

```go
package retry

import (
    "context"
    "time"
)

// Do executes an operation with retry
func Do(ctx context.Context, operation Operation, opts ...Option) error

// DoWithResult executes and returns a result
func DoWithResult[T any](ctx context.Context, operation OperationWithResult[T], opts ...Option) (T, error)

// Operation is a retriable operation
type Operation func(ctx context.Context) error

// OperationWithResult is a retriable operation with result
type OperationWithResult[T any] func(ctx context.Context) (T, error)

// Backoff calculates delay between retries
type Backoff interface {
    // Next returns the delay for attempt n (0-indexed)
    Next(attempt int) time.Duration

    // Reset resets the backoff state
    Reset()
}

// RetryIf determines if an error should be retried
type RetryIf func(error) bool

// OnRetry is called before each retry
type OnRetry func(attempt int, err error, delay time.Duration)
```

## Configuration

```go
// Config holds retry configuration
type Config struct {
    // Maximum attempts (including initial)
    MaxAttempts int

    // Backoff strategy
    Backoff Backoff

    // Condition for retry
    RetryIf RetryIf

    // Callback on retry
    OnRetry OnRetry

    // Maximum total duration
    MaxDuration time.Duration

    // Last error only (don't wrap)
    LastErrorOnly bool
}

// Option configures retry behavior
type Option func(*Config)

// WithMaxAttempts sets maximum attempts
func WithMaxAttempts(n int) Option

// WithBackoff sets the backoff strategy
func WithBackoff(b Backoff) Option

// WithRetryIf sets the retry condition
func WithRetryIf(fn RetryIf) Option

// WithOnRetry sets the retry callback
func WithOnRetry(fn OnRetry) Option

// WithMaxDuration sets maximum total duration
func WithMaxDuration(d time.Duration) Option

// WithContext uses context for cancellation
func WithContext(ctx context.Context) Option
```

## Backoff Strategies

```go
// Constant backoff with fixed delay
type ConstantBackoff struct {
    Delay time.Duration
}

func NewConstantBackoff(delay time.Duration) *ConstantBackoff

// Linear backoff with increasing delay
type LinearBackoff struct {
    Initial   time.Duration
    Increment time.Duration
    Max       time.Duration
}

func NewLinearBackoff(initial, increment, max time.Duration) *LinearBackoff

// Exponential backoff with doubling delay
type ExponentialBackoff struct {
    Initial    time.Duration
    Multiplier float64
    Max        time.Duration
}

func NewExponentialBackoff(initial time.Duration, multiplier float64, max time.Duration) *ExponentialBackoff

// Default exponential: 100ms, 2x multiplier, 30s max
func DefaultExponentialBackoff() *ExponentialBackoff

// Fibonacci backoff following fibonacci sequence
type FibonacciBackoff struct {
    Initial time.Duration
    Max     time.Duration
}

func NewFibonacciBackoff(initial, max time.Duration) *FibonacciBackoff
```

## Jitter

```go
// Jitter adds randomness to backoff
type Jitter interface {
    Apply(delay time.Duration) time.Duration
}

// FullJitter: delay = random(0, delay)
type FullJitter struct{}

// EqualJitter: delay = delay/2 + random(0, delay/2)
type EqualJitter struct{}

// DecorrelatedJitter: delay = min(max, random(base, delay * 3))
type DecorrelatedJitter struct {
    Base time.Duration
    Max  time.Duration
}

// WithJitter wraps a backoff with jitter
func WithJitter(backoff Backoff, jitter Jitter) Backoff
```

## Usage Examples

### Basic Retry

```go
package main

import (
    "context"
    "errors"
    "github.com/user/core-backend/pkg/retry"
)

func main() {
    ctx := context.Background()

    err := retry.Do(ctx, func(ctx context.Context) error {
        return callExternalService()
    },
        retry.WithMaxAttempts(3),
        retry.WithBackoff(retry.NewExponentialBackoff(
            100*time.Millisecond, // Initial delay
            2.0,                   // Multiplier
            30*time.Second,        // Max delay
        )),
    )

    if err != nil {
        log.Printf("Failed after retries: %v", err)
    }
}
```

### With Result

```go
func main() {
    ctx := context.Background()

    result, err := retry.DoWithResult(ctx, func(ctx context.Context) (string, error) {
        return fetchData()
    },
        retry.WithMaxAttempts(5),
    )

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Result:", result)
}
```

### Conditional Retry

```go
func main() {
    ctx := context.Background()

    // Only retry on specific errors
    err := retry.Do(ctx, operation,
        retry.WithMaxAttempts(3),
        retry.WithRetryIf(func(err error) bool {
            // Retry on timeout or temporary errors
            var netErr net.Error
            if errors.As(err, &netErr) {
                return netErr.Timeout() || netErr.Temporary()
            }
            // Don't retry on permanent errors
            return !errors.Is(err, ErrNotFound)
        }),
    )
}
```

### With Jitter

```go
func main() {
    ctx := context.Background()

    backoff := retry.WithJitter(
        retry.NewExponentialBackoff(100*time.Millisecond, 2.0, 30*time.Second),
        &retry.FullJitter{},
    )

    err := retry.Do(ctx, operation,
        retry.WithBackoff(backoff),
        retry.WithMaxAttempts(5),
    )
}
```

### With Callbacks

```go
func main() {
    ctx := context.Background()

    err := retry.Do(ctx, operation,
        retry.WithMaxAttempts(5),
        retry.WithOnRetry(func(attempt int, err error, delay time.Duration) {
            log.Printf("Attempt %d failed: %v. Retrying in %v", attempt, err, delay)
        }),
    )
}
```

### HTTP Client Retry

```go
func fetchWithRetry(url string) (*http.Response, error) {
    ctx := context.Background()

    return retry.DoWithResult(ctx, func(ctx context.Context) (*http.Response, error) {
        req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return nil, err
        }

        // Retry on 5xx errors
        if resp.StatusCode >= 500 {
            resp.Body.Close()
            return nil, fmt.Errorf("server error: %d", resp.StatusCode)
        }

        return resp, nil
    },
        retry.WithMaxAttempts(3),
        retry.WithRetryIf(func(err error) bool {
            // Retry on network errors and 5xx
            return true
        }),
    )
}
```

### Database Retry

```go
func executeWithRetry(db *sql.DB, query string) error {
    ctx := context.Background()

    return retry.Do(ctx, func(ctx context.Context) error {
        _, err := db.ExecContext(ctx, query)
        return err
    },
        retry.WithMaxAttempts(3),
        retry.WithBackoff(retry.NewExponentialBackoff(
            50*time.Millisecond,
            2.0,
            1*time.Second,
        )),
        retry.WithRetryIf(func(err error) bool {
            // Retry on deadlock or serialization failure
            return isRetryableDBError(err)
        }),
    )
}
```

### With Timeout

```go
func main() {
    // Total timeout for all retries
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err := retry.Do(ctx, operation,
        retry.WithMaxAttempts(10), // May not reach 10 if timeout
    )

    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Timed out")
    }
}
```

### Permanent Errors

```go
// PermanentError wraps an error that should not be retried
type PermanentError struct {
    Err error
}

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

// Permanent marks an error as non-retriable
func Permanent(err error) error {
    return &PermanentError{Err: err}
}

// Usage
func main() {
    err := retry.Do(ctx, func(ctx context.Context) error {
        resp, err := callAPI()
        if err != nil {
            return err
        }

        if resp.StatusCode == 400 {
            // Don't retry client errors
            return retry.Permanent(fmt.Errorf("bad request"))
        }

        return nil
    })
}
```

### Retry All Errors

```go
// Default retry condition (retry all except permanent)
func DefaultRetryIf(err error) bool {
    var permanent *PermanentError
    return !errors.As(err, &permanent)
}

// Retry everything
func AlwaysRetry(err error) bool {
    return true
}

// Never retry
func NeverRetry(err error) bool {
    return false
}
```

## Error Handling

```go
// RetryError contains all errors from attempts
type RetryError struct {
    Attempts []AttemptError
}

type AttemptError struct {
    Attempt int
    Err     error
    At      time.Time
}

func (e *RetryError) Error() string
func (e *RetryError) Unwrap() error // Returns last error

// IsRetryError checks if error is from retry
func IsRetryError(err error) bool
```

## Dependencies

- **Required:** None (pure Go)

## Implementation Phases

### Phase 1: Core Retry
1. Basic Do function
2. MaxAttempts configuration
3. Simple backoff

### Phase 2: Backoff Strategies
1. Constant backoff
2. Linear backoff
3. Exponential backoff

### Phase 3: Jitter
1. Full jitter
2. Equal jitter
3. Decorrelated jitter

### Phase 4: Advanced Features
1. RetryIf conditions
2. OnRetry callbacks
3. Permanent errors
4. Generics support

### Phase 5: Documentation
1. README
2. Examples
3. Best practices
