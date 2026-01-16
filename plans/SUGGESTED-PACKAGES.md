# Suggested New Packages for core-backend

This document outlines new packages that would complement the existing `core-backend` library. Each package follows the established patterns:

- **Interface-based design** for extensibility
- **Functional options pattern** for configuration
- **Environment-driven configuration** with sensible defaults
- **Minimal external dependencies** (standard library first)
- **80%+ test coverage** with unit, integration, and benchmark tests

## Existing Packages

| Package | Purpose |
|---------|---------|
| `pkg/auth` | Authentication & Authorization with JWT, RBAC, and account protection |
| `pkg/httpclient` | Production HTTP client with retry and circuit breaker |
| `pkg/postgres` | PostgreSQL database client with connection pooling |
| `pkg/server` | Unified gRPC & HTTP server with gRPC-Gateway |

## Suggested New Packages

### 1. pkg/cache - Caching Layer

**Purpose:** Unified caching interface supporting Redis and in-memory stores.

**Why:** Most production applications need caching for performance. A unified interface allows swapping backends without code changes.

**Plan:** [PLAN-cache.md](./PLAN-cache.md)

---

### 2. pkg/queue - Message Queue Client

**Purpose:** Asynchronous message processing with support for Kafka, RabbitMQ, and NATS.

**Why:** Event-driven architectures and microservices communication require reliable message queuing.

**Plan:** [PLAN-queue.md](./PLAN-queue.md)

---

### 3. pkg/storage - File Storage Abstraction

**Purpose:** Unified file storage interface supporting S3, GCS, Azure Blob, and local filesystem.

**Why:** Applications often need to store and retrieve files. A unified interface enables cloud-agnostic implementations.

**Plan:** [PLAN-storage.md](./PLAN-storage.md)

---

### 4. pkg/email - Email Service

**Purpose:** Email sending with template support, multiple providers (SMTP, SendGrid, SES).

**Why:** Nearly every application needs to send transactional emails (welcome, password reset, notifications).

**Plan:** [PLAN-email.md](./PLAN-email.md)

---

### 5. pkg/observability - Tracing & Metrics

**Purpose:** Distributed tracing and metrics collection using OpenTelemetry.

**Why:** Production systems require observability for debugging, performance monitoring, and alerting.

**Plan:** [PLAN-observability.md](./PLAN-observability.md)

---

### 6. pkg/scheduler - Background Job Processing

**Purpose:** Background job scheduling and processing with cron support.

**Why:** Applications need to run periodic tasks (cleanup, reports, sync) and process async jobs.

**Plan:** [PLAN-scheduler.md](./PLAN-scheduler.md)

---

## Implementation Priority

Recommended implementation order based on common usage patterns:

| Priority | Package | Reason |
|----------|---------|--------|
| 1 | `cache` | Most applications need caching; pairs well with `postgres` |
| 2 | `observability` | Critical for production debugging; integrates with all packages |
| 3 | `email` | Common requirement for auth flows (password reset) |
| 4 | `queue` | Essential for microservices and async processing |
| 5 | `scheduler` | Background processing is a common need |
| 6 | `storage` | File handling is project-specific |

## Package Independence

Following the existing pattern, each new package will be **completely independent**:

- No imports from other `core-backend` packages
- Can be used standalone or together
- Interfaces allow custom implementations
- All dependencies are optional (pluggable)

## Integration Examples

While packages are independent, they work well together:

```go
// Example: Rate limiting with Redis cache
limiter := ratelimit.New(cache.NewRedisStore(redisClient))
server := server.New(server.WithRateLimiter(limiter))

// Example: Async email sending via queue
queue.Publish("emails", email.NewWelcomeEmail(user))

// Example: Scheduled cache cleanup
scheduler.Every(1).Hour().Do(func() {
    cache.DeleteExpired()
})

// Example: Traced database queries
ctx := observability.StartSpan(ctx, "db.query")
rows, err := postgres.Query(ctx, query)
observability.EndSpan(ctx)
```
