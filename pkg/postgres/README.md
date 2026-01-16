# Postgres Package

The `pkg/postgres` module provides a plug-and-play PostgreSQL client with connection pooling, transaction support, and production-ready defaults. It uses [pgx v5](https://github.com/jackc/pgx) as the underlying driver.

## Features

- **Connection pooling** with configurable limits
- **Transaction support** with automatic rollback on error
- **Environment-based configuration** with sensible defaults
- **Health checks** for service integrations
- **Error helpers** for constraint violations
- **Generic row scanning** utilities
- **Schema support** for multi-tenant applications

## Configuration

Call `postgres.LoadConfig()` at startup to gather values from the environment:

| Env var | Purpose | Default |
|---------|---------|---------|
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Username | **required** |
| `POSTGRES_PASSWORD` | Password | **required** |
| `POSTGRES_DATABASE` | Database name | **required** |
| `POSTGRES_SCHEMA` | Database schema | `public` |
| `POSTGRES_SSL_MODE` | SSL mode | `prefer` |
| `POSTGRES_MAX_CONNS` | Maximum connections | `25` |
| `POSTGRES_MIN_CONNS` | Minimum connections | `5` |
| `POSTGRES_MAX_CONN_LIFETIME` | Max connection lifetime | `1h` |
| `POSTGRES_MAX_CONN_IDLE_TIME` | Max idle time | `30m` |
| `POSTGRES_CONNECT_TIMEOUT` | Connection timeout | `10s` |
| `POSTGRES_QUERY_TIMEOUT` | Default query timeout | `30s` |

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/postgres"
)

func main() {
    // Load config from environment
    cfg, err := postgres.LoadConfig()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    // Create client
    client, err := postgres.New(*cfg)
    if err != nil {
        log.Fatalf("connect: %v", err)
    }
    defer client.Close()

    // Query
    ctx := context.Background()
    rows, err := client.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
    if err != nil {
        log.Fatalf("query: %v", err)
    }

    // Scan results
    type User struct {
        ID   int    `db:"id"`
        Name string `db:"name"`
    }
    users, err := postgres.ScanAll[User](rows)
    if err != nil {
        log.Fatalf("scan: %v", err)
    }

    log.Printf("Found %d users", len(users))
}
```

## Custom Environment Variables

If you need to use custom env var names (e.g., `MYAPP_DB_*`):

```go
cfg := postgres.Config{
    Host:     os.Getenv("MYAPP_DB_HOST"),
    Port:     5432,
    User:     os.Getenv("MYAPP_DB_USER"),
    Password: os.Getenv("MYAPP_DB_PASSWORD"),
    Database: os.Getenv("MYAPP_DB_NAME"),
    Schema:   os.Getenv("MYAPP_DB_SCHEMA"),
}

client, err := postgres.New(cfg)
```

Or use a connection URL:

```go
client, err := postgres.NewFromURL(os.Getenv("DATABASE_URL"))
```

## Transactions

```go
err := client.Transaction(ctx, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
    if err != nil {
        return err // Automatic rollback
    }

    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - 100 WHERE user_id = $1", 1)
    return err // Commit on nil, rollback on error
})
```

## Error Handling

```go
_, err := client.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", email)
if err != nil {
    if postgres.IsUniqueViolation(err) {
        return fmt.Errorf("email already exists")
    }
    if postgres.IsForeignKeyViolation(err) {
        return fmt.Errorf("referenced record not found")
    }
    return err
}
```

## Health Checks

```go
status := client.Health(ctx)
if !status.Healthy {
    log.Printf("Database unhealthy: %s", status.Message)
}
log.Printf("Latency: %v, Connections: %d/%d",
    status.Latency, status.ActiveConns, status.TotalConns)
```

## Testing

Run unit tests:
```bash
go test ./pkg/postgres/... -v -short
```

Run integration tests (requires PostgreSQL):
```bash
docker run -d --name postgres-test \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16-alpine

POSTGRES_USER=test POSTGRES_PASSWORD=test POSTGRES_DATABASE=testdb \
  go test ./pkg/postgres/... -v -run Integration
```
