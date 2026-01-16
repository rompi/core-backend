# PostgreSQL Plug and Play Package Implementation Plan

Create a production-ready, plug-and-play PostgreSQL package (`pkg/postgres`) following the established patterns from `pkg/auth` and `pkg/httpclient`.

## Proposed Changes

### PostgreSQL Package (`pkg/postgres`)

#### [NEW] [config.go](file:./pkg/postgres/config.go)

Configuration management with environment variable support:

| Env var | Purpose | Default |
|---------|---------|---------|
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Username | **required** |
| `POSTGRES_PASSWORD` | Password | **required** |
| `POSTGRES_DATABASE` | Database name | **required** |
| `POSTGRES_SSL_MODE` | SSL mode | `prefer` |
| `POSTGRES_MAX_CONNS` | Maximum connections | `25` |
| `POSTGRES_MIN_CONNS` | Minimum connections | `5` |
| `POSTGRES_MAX_CONN_LIFETIME` | Max connection lifetime | `1h` |
| `POSTGRES_MAX_CONN_IDLE_TIME` | Max idle time | `30m` |
| `POSTGRES_CONNECT_TIMEOUT` | Connection timeout | `10s` |
| `POSTGRES_QUERY_TIMEOUT` | Default query timeout | `30s` |
| `POSTGRES_SCHEMA` | Database schema | `public` |

```go
type Config struct {
    Host               string
    Port               int
    User               string
    Password           string
    Database           string
    Schema             string        // Database schema (default: "public")
    SSLMode            string
    MaxConns           int32
    MinConns           int32
    MaxConnLifetime    time.Duration
    MaxConnIdleTime    time.Duration
    ConnectTimeout     time.Duration
    QueryTimeout       time.Duration
}

func LoadConfig() (*Config, error)
func (c *Config) Validate() error
func (c *Config) ConnectionString() string
```

---

#### [NEW] [errors.go](file:./pkg/postgres/errors.go)

Custom error types for PostgreSQL operations:

```go
var (
    ErrConnectionFailed   = errors.New("postgres: connection failed")
    ErrQueryFailed        = errors.New("postgres: query failed")
    ErrNoRows             = errors.New("postgres: no rows")
    ErrConstraintViolation = errors.New("postgres: constraint violation")
    ErrDuplicateKey       = errors.New("postgres: duplicate key")
    ErrInvalidConfig      = errors.New("postgres: invalid configuration")
    ErrTimeout            = errors.New("postgres: timeout")
    ErrPoolExhausted      = errors.New("postgres: connection pool exhausted")
)

// IsUniqueViolation checks if error is a unique constraint violation
func IsUniqueViolation(err error) bool

// IsForeignKeyViolation checks if error is a foreign key violation  
func IsForeignKeyViolation(err error) bool
```

---

#### [NEW] [client.go](file:./pkg/postgres/client.go)

Main client with connection pooling using pgxpool:

```go
type Client struct {
    pool   *pgxpool.Pool
    config *Config
    logger Logger
}

// New creates a new PostgreSQL client
func New(cfg Config, opts ...Option) (*Client, error)

// NewFromURL creates a client from a connection URL
func NewFromURL(url string, opts ...Option) (*Client, error)

// Pool returns the underlying connection pool
func (c *Client) Pool() *pgxpool.Pool

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

// QueryRow executes a query that returns at most one row
func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

// Exec executes a query that doesn't return rows
func (c *Client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

// Transaction executes a function within a transaction
func (c *Client) Transaction(ctx context.Context, fn func(tx pgx.Tx) error) error

// TransactionWithOptions executes with custom isolation level
func (c *Client) TransactionWithOptions(ctx context.Context, opts pgx.TxOptions, fn func(tx pgx.Tx) error) error

// Ping checks if the database is reachable
func (c *Client) Ping(ctx context.Context) error

// Close closes the connection pool
func (c *Client) Close()

// Stats returns connection pool statistics
func (c *Client) Stats() *PoolStats
```

---

#### [NEW] [options.go](file:./pkg/postgres/options.go)

Functional options pattern:

```go
type Option func(*Client)

func WithLogger(logger Logger) Option
func WithQueryHook(hook QueryHook) Option
func WithTracer(tracer Tracer) Option
```

---

#### [NEW] [logger.go](file:./pkg/postgres/logger.go)

Logger interface (matches `pkg/httpclient` pattern):

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
}

type NoopLogger struct{}
func NewNoopLogger() *NoopLogger
```

---

#### [NEW] [health.go](file:./pkg/postgres/health.go)

Health check for service integrations:

```go
type HealthStatus struct {
    Healthy     bool
    Message     string
    Latency     time.Duration
    ActiveConns int32
    IdleConns   int32
    TotalConns  int32
}

func (c *Client) Health(ctx context.Context) HealthStatus
```

---

#### [NEW] [scanner.go](file:./pkg/postgres/scanner.go)

Helper utilities for row scanning:

```go
// ScanOne scans a single row into a struct
func ScanOne[T any](rows pgx.Rows) (*T, error)

// ScanAll scans all rows into a slice of structs  
func ScanAll[T any](rows pgx.Rows) ([]T, error)

// ScanMap scans a row into a map
func ScanMap(rows pgx.Rows) (map[string]any, error)
```

---

#### [NEW] [README.md](file:./pkg/postgres/README.md)

Documentation following `pkg/auth/README.md` pattern:
- Design overview
- Configuration reference
- Usage examples
- Error handling guide
- Best practices

---

### Examples

#### [NEW] [examples/basic/main.go](file:./pkg/postgres/examples/basic/main.go)

Basic connection and query example.

#### [NEW] [examples/transactions/main.go](file:./pkg/postgres/examples/transactions/main.go)

Transaction usage example.

#### [NEW] [examples/health-check/main.go](file:./pkg/postgres/examples/health-check/main.go)

Health check integration example.

#### [NEW] [examples/custom-env/main.go](file:./pkg/postgres/examples/custom-env/main.go)

Connect with custom environment variable prefix (e.g., `MYAPP_DB_*` instead of `POSTGRES_*`):

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rompi/core-backend/pkg/postgres"
)

func main() {
    // Option 1: Create config manually from custom env vars
    cfg := postgres.Config{
        Host:     getEnv("MYAPP_DB_HOST", "localhost"),
        Port:     5432,
        User:     os.Getenv("MYAPP_DB_USER"),
        Password: os.Getenv("MYAPP_DB_PASSWORD"),
        Database: os.Getenv("MYAPP_DB_NAME"),
        SSLMode:  getEnv("MYAPP_DB_SSL_MODE", "prefer"),
    }

    if err := cfg.Validate(); err != nil {
        log.Fatalf("invalid config: %v", err)
    }

    client, err := postgres.New(cfg)
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
    defer client.Close()

    // Option 2: Use connection URL from custom env var
    // url := os.Getenv("MYAPP_DATABASE_URL")
    // client, err := postgres.NewFromURL(url)

    ctx := context.Background()
    if err := client.Ping(ctx); err != nil {
        log.Fatalf("failed to ping: %v", err)
    }

    fmt.Println("Connected successfully with custom env vars!")
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

### Dependencies

#### [MODIFY] [go.mod](file:./go.mod)

Add pgx v5 driver:
```diff
require (
    github.com/golang-jwt/jwt/v5 v5.3.0
    github.com/google/uuid v1.6.0
    golang.org/x/crypto v0.45.0
    modernc.org/sqlite v1.40.1
+   github.com/jackc/pgx/v5 v5.7.4
)
```

---

## Verification Plan

### Unit Tests

Run unit tests to verify configuration, error handling, and helpers:

```bash
go test ./pkg/postgres/... -v -short
```

**Test files to create:**
- `config_test.go` - Config loading and validation
- `errors_test.go` - Error type checking helpers
- `client_test.go` - Client creation (mocked pool)
- `scanner_test.go` - Row scanning utilities

### Integration Tests

Integration tests require a running PostgreSQL instance:

```bash
# Start PostgreSQL
docker run -d --name postgres-test \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16-alpine

# Run tests
POSTGRES_USER=test POSTGRES_PASSWORD=test POSTGRES_DATABASE=testdb \
  go test ./pkg/postgres/... -v -run Integration

# Cleanup
docker stop postgres-test && docker rm postgres-test
```

### Lint Check

```bash
make lint
```
