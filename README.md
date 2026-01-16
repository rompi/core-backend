# core-backend

A production-ready Go backend library providing plug-and-play packages for building scalable applications with authentication, HTTP capabilities, database access, and unified server management.

## Features

- **Modular Design** - Independent packages that can be adopted incrementally
- **Production-Ready** - Enterprise-grade patterns with 80%+ test coverage
- **Minimal Dependencies** - Standard library first approach
- **Interface-Based** - Extensible through clean abstractions
- **Well Documented** - Comprehensive examples for every package

## Packages

### `pkg/auth` - Authentication & Authorization

Enterprise-grade authentication service with:
- User registration and secure login with JWT tokens
- Account lockout and brute-force protection
- Password reset with secure tokens
- Role-based access control (RBAC)
- API key validation for M2M authentication
- Audit logging and rate limiting
- HTTP middleware helpers

[View auth documentation](pkg/auth/README.md)

### `pkg/httpclient` - Production HTTP Client

Robust HTTP client library with:
- Fluent API for building requests
- Automatic retry with exponential backoff
- Circuit breaker pattern
- Middleware system (auth, logging, custom headers)
- Zero external dependencies

[View httpclient documentation](pkg/httpclient/README.md)

### `pkg/postgres` - PostgreSQL Database Client

PostgreSQL driver wrapper with:
- Connection pooling with configurable limits
- Transaction support with automatic rollback
- Environment-based configuration
- Health checks and error helpers
- Schema support for multi-tenant apps

[View postgres documentation](pkg/postgres/README.md)

### `pkg/server` - Unified gRPC & HTTP Server

Unified server managing both gRPC and HTTP/REST APIs:
- Proto-first architecture
- gRPC-Gateway for automatic REST endpoints
- Built-in interceptors (auth, logging, recovery, rate limiting)
- Graceful shutdown support

[View server documentation](pkg/server/PLAN.md)

## Installation

```bash
go get github.com/rompi/core-backend
```

Import the packages you need:

```go
import (
    "github.com/rompi/core-backend/pkg/auth"
    "github.com/rompi/core-backend/pkg/httpclient"
    "github.com/rompi/core-backend/pkg/postgres"
)
```

## Quick Start

### Authentication

```go
// Create auth service with your repository implementations
cfg := auth.DefaultConfig()
cfg.JWTSecret = "your-secret-key"

service := auth.NewService(cfg, userRepo, sessionRepo, roleRepo, auditRepo)

// Register a user
user, err := service.Register(ctx, "user@example.com", "securepassword123")

// Login and get JWT token
token, err := service.Login(ctx, "user@example.com", "securepassword123")
```

### HTTP Client

```go
// Create client with retry and circuit breaker
client := httpclient.New(
    httpclient.WithBaseURL("https://api.example.com"),
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithRetry(3, time.Second, 30*time.Second),
)

// Make requests with fluent API
resp, err := client.Get(ctx, "/users").
    WithHeader("Accept", "application/json").
    Do()
```

### PostgreSQL

```go
// Connect using environment variables
db, err := postgres.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Execute queries
rows, err := db.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
```

## Configuration

All packages use environment variables for configuration:

### Auth

| Variable | Description | Default |
|----------|-------------|---------|
| `AUTH_JWT_SECRET` | JWT signing secret | Required |
| `AUTH_JWT_EXPIRATION` | Token lifetime | `24h` |
| `AUTH_PASSWORD_MIN_LENGTH` | Minimum password length | `8` |
| `AUTH_MAX_FAILED_ATTEMPTS` | Lockout threshold | `5` |

### PostgreSQL

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | Required |
| `POSTGRES_PASSWORD` | Database password | Required |
| `POSTGRES_DATABASE` | Database name | Required |
| `POSTGRES_SSL_MODE` | SSL mode | `prefer` |

## Development

### Prerequisites

- Go 1.25+
- Make

### Commands

```bash
# Run all tests
make test

# Run tests with coverage
make cover

# Run linter
make lint

# Format code
make fmt

# Build all packages
make build
```

### Project Structure

```
core-backend/
├── pkg/
│   ├── auth/           # Authentication package
│   ├── httpclient/     # HTTP client package
│   ├── postgres/       # PostgreSQL client package
│   └── server/         # gRPC/HTTP server package
├── docs/               # Additional documentation
├── coding-guidelines.md
├── CHANGELOG.md
└── Makefile
```

## Documentation

- [API Reference](docs/API.md)
- [Examples](docs/EXAMPLES.md)
- [Integration Guide](docs/INTEGRATION.md)
- [Coding Guidelines](coding-guidelines.md)
- [Changelog](CHANGELOG.md)

## License

See [LICENSE](LICENSE) for details.
