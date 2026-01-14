# httpclient

Production-grade HTTP client for Go with built-in retry, circuit breaking, and middleware support.

## Features

- 🚀 **Fluent API** for building HTTP requests
- 🔄 **Automatic retry** with exponential backoff for transient failures
- 🛡️ **Circuit breaker** pattern to prevent cascading failures
- 🔌 **Middleware system** for request/response interception
- 📝 **Structured logging** with pluggable logger interface
- 🎯 **JSON helpers** for easy encoding/decoding
- ⚡ **Context-aware** with built-in cancellation and timeout support
- 🧪 **Comprehensive tests** with 80%+ coverage
- 📦 **Zero dependencies** beyond Go standard library

## Installation

```bash
go get github.com/rompi/core-backend/pkg/httpclient
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rompi/core-backend/pkg/httpclient"
)

func main() {
    // Create a client with defaults
    client := httpclient.NewDefault("https://api.example.com")

    // Make a GET request
    resp, err := client.Get(context.Background(), "/users/123").Do()
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Decode JSON response
    var user struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }
    if err := resp.JSON(&user); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("User: %+v\n", user)
}
```

## Configuration

Create a client with custom configuration:

```go
client, err := httpclient.New(httpclient.Config{
    BaseURL:      "https://api.example.com",
    Timeout:      30 * time.Second,
    MaxRetries:   3,
    RetryWaitMin: 1 * time.Second,
    RetryWaitMax: 30 * time.Second,
    Logger:       myLogger, // Optional: implement httpclient.Logger interface
})
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `BaseURL` | `string` | **(required)** | Base URL for all requests |
| `Timeout` | `time.Duration` | `30s` | Maximum duration for a request |
| `MaxRetries` | `int` | `3` | Maximum number of retry attempts |
| `RetryWaitMin` | `time.Duration` | `1s` | Minimum wait time between retries |
| `RetryWaitMax` | `time.Duration` | `30s` | Maximum wait time between retries |
| `CircuitBreaker` | `*CircuitBreakerConfig` | `nil` | Circuit breaker configuration |
| `Logger` | `Logger` | noop logger | Logger implementation |
| `Transport` | `http.RoundTripper` | `http.DefaultTransport` | HTTP transport |
| `FollowRedirects` | `bool` | `true` | Whether to follow HTTP redirects |

## HTTP Methods

```go
// GET request
resp, err := client.Get(ctx, "/users").Do()

// POST with JSON body
resp, err := client.Post(ctx, "/users").
    JSON(map[string]string{"name": "John"}).
    Do()

// PUT request
resp, err := client.Put(ctx, "/users/123").
    JSON(updateData).
    Do()

// PATCH request
resp, err := client.Patch(ctx, "/users/123").
    JSON(partialUpdate).
    Do()

// DELETE request
resp, err := client.Delete(ctx, "/users/123").Do()
```

## Request Building

The fluent API allows you to build complex requests:

```go
resp, err := client.Get(ctx, "/search").
    Header("Authorization", "Bearer token").
    Header("X-Custom-Header", "value").
    Query("q", "golang").
    Query("limit", "10").
    Do()
```

Or use batch methods:

```go
resp, err := client.Post(ctx, "/api/data").
    Headers(map[string]string{
        "Authorization": "Bearer token",
        "Content-Type":  "application/json",
    }).
    QueryParams(map[string]string{
        "filter": "active",
        "sort":   "name",
    }).
    JSON(requestBody).
    Do()
```

## Response Helpers

```go
resp, err := client.Get(ctx, "/users/123").Do()
if err != nil {
    return err
}

// Check status
if resp.IsSuccess() {
    // 2xx status
}
if resp.IsClientError() {
    // 4xx status
}
if resp.IsServerError() {
    // 5xx status
}

// Decode JSON
var user User
if err := resp.JSON(&user); err != nil {
    return err
}

// Get as string
body, err := resp.String()

// Get as bytes
data, err := resp.Bytes()
```

## Middleware

Add middleware to intercept and modify requests/responses:

```go
client := httpclient.NewDefault("https://api.example.com")

// Add authentication
client.Use(httpclient.AuthBearerMiddleware("your-token"))

// Add custom user agent
client.Use(httpclient.UserAgentMiddleware("MyApp/1.0"))

// Add logging
client.Use(httpclient.LoggingMiddleware(logger))

// Add custom headers
client.Use(httpclient.HeaderMiddleware(map[string]string{
    "X-API-Version": "v1",
    "X-Request-ID":  uuid.New().String(),
}))
```

### Built-in Middleware

- `LoggingMiddleware(logger)` - Logs requests and responses
- `AuthBearerMiddleware(token)` - Adds Bearer token authorization
- `AuthAPIKeyMiddleware(headerName, apiKey)` - Adds API key header
- `UserAgentMiddleware(userAgent)` - Sets User-Agent header
- `HeaderMiddleware(headers)` - Adds custom headers

### Custom Middleware

```go
func CustomMiddleware() httpclient.Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return httpclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            // Modify request
            req.Header.Set("X-Custom", "value")

            // Execute request
            resp, err := next.RoundTrip(req)

            // Modify response
            if resp != nil {
                // ...
            }

            return resp, err
        })
    }
}

client.Use(CustomMiddleware())
```

## Automatic Retry

The client automatically retries failed requests for transient errors:

- Network errors (connection refused, timeout, etc.)
- 5xx server errors
- 429 Too Many Requests

Retries use exponential backoff with jitter to prevent thundering herd.

```go
client, err := httpclient.New(httpclient.Config{
    BaseURL:      "https://api.example.com",
    MaxRetries:   5,                  // Retry up to 5 times
    RetryWaitMin: 2 * time.Second,    // Start with 2s wait
    RetryWaitMax: 60 * time.Second,   // Cap at 60s wait
})
```

**What gets retried:**
- ✅ Network failures (connection refused, DNS errors, timeouts)
- ✅ 500 Internal Server Error
- ✅ 502 Bad Gateway
- ✅ 503 Service Unavailable
- ✅ 504 Gateway Timeout
- ✅ 429 Too Many Requests
- ❌ 4xx client errors (except 429)
- ❌ 2xx successful responses

## Circuit Breaker

Prevent cascading failures with the circuit breaker pattern:

```go
client, err := httpclient.New(httpclient.Config{
    BaseURL: "https://api.example.com",
    CircuitBreaker: &httpclient.CircuitBreakerConfig{
        MaxRequests: 5,                   // Max requests in half-open state
        Interval:    10 * time.Second,    // Reset interval for closed state
        Timeout:     60 * time.Second,    // Time to wait before half-open
        ReadyToTrip: func(counts httpclient.Counts) bool {
            // Open circuit after 5 consecutive failures
            return counts.ConsecutiveFailures >= 5
        },
    },
})
```

### Circuit Breaker States

1. **Closed** - Normal operation, requests flow through
2. **Open** - Too many failures, requests are rejected immediately
3. **Half-Open** - Testing if service recovered, limited requests allowed

## Error Handling

```go
resp, err := client.Get(ctx, "/api/data").Do()
if err != nil {
    // Check for specific errors
    if errors.Is(err, httpclient.ErrTimeout) {
        // Handle timeout
    }
    if errors.Is(err, httpclient.ErrCircuitOpen) {
        // Handle circuit breaker open
    }
    if errors.Is(err, httpclient.ErrMaxRetriesExceeded) {
        // Handle max retries
    }

    // Check for HTTP error
    var httpErr *httpclient.Error
    if errors.As(err, &httpErr) {
        fmt.Printf("HTTP error: %d - %s\n", httpErr.StatusCode, httpErr.Message)
    }

    return err
}
```

## Context and Cancellation

All requests support context cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.Get(ctx, "/slow-endpoint").Do()

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(1 * time.Second)
    cancel() // Cancel after 1 second
}()

resp, err := client.Get(ctx, "/endpoint").Do()
```

## Logging

Implement the `Logger` interface to use your preferred logging framework:

```go
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}
```

Example with slog:

```go
type SlogAdapter struct {
    logger *slog.Logger
}

func (a *SlogAdapter) Debug(msg string, keysAndValues ...interface{}) {
    a.logger.Debug(msg, keysAndValues...)
}

func (a *SlogAdapter) Info(msg string, keysAndValues ...interface{}) {
    a.logger.Info(msg, keysAndValues...)
}

func (a *SlogAdapter) Warn(msg string, keysAndValues ...interface{}) {
    a.logger.Warn(msg, keysAndValues...)
}

func (a *SlogAdapter) Error(msg string, keysAndValues ...interface{}) {
    a.logger.Error(msg, keysAndValues...)
}

// Use with httpclient
client, err := httpclient.New(httpclient.Config{
    BaseURL: "https://api.example.com",
    Logger:  &SlogAdapter{logger: slog.Default()},
})
```

## Testing

The package includes comprehensive tests with 80%+ coverage. Run tests with:

```bash
# Run all tests
go test ./pkg/httpclient/...

# Run with coverage
go test -cover ./pkg/httpclient/...

# Run with race detector
go test -race ./pkg/httpclient/...
```

## Examples

See the [examples/](examples/) directory for complete examples:

- [basic/](examples/basic/) - Simple GET/POST requests
- [middleware/](examples/middleware/) - Custom middleware, auth, logging
- [advanced/](examples/advanced/) - Circuit breaker, retry policies, error handling

## API Reference

Full API documentation is available on [GoDoc](https://pkg.go.dev/github.com/rompi/core-backend/pkg/httpclient).

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`go test ./...`)
- Code is formatted (`gofmt`)
- Coverage remains above 80% (`go test -cover`)
- Follow existing code style

## Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for version history and changes.
