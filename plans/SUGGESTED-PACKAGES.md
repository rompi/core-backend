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

---

## Infrastructure Packages

### Core Infrastructure

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 1 | `pkg/cache` | Redis & in-memory caching | [PLAN-cache.md](./PLAN-cache.md) |
| 2 | `pkg/queue` | Message queue (Kafka, RabbitMQ, NATS) | [PLAN-queue.md](./PLAN-queue.md) |
| 3 | `pkg/storage` | File storage (S3, GCS, Azure, local) | [PLAN-storage.md](./PLAN-storage.md) |
| 4 | `pkg/scheduler` | Background jobs with cron support | [PLAN-scheduler.md](./PLAN-scheduler.md) |
| 5 | `pkg/pubsub` | Pub/Sub for event broadcasting | [PLAN-pubsub.md](./PLAN-pubsub.md) |

### Resilience & Reliability

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 6 | `pkg/ratelimit` | Rate limiting (token bucket, sliding window) | [PLAN-ratelimit.md](./PLAN-ratelimit.md) |
| 7 | `pkg/circuitbreaker` | Circuit breaker for fault tolerance | [PLAN-circuitbreaker.md](./PLAN-circuitbreaker.md) |
| 8 | `pkg/retry` | Retry with exponential backoff | [PLAN-retry.md](./PLAN-retry.md) |
| 9 | `pkg/lock` | Distributed locking (Redis, PostgreSQL) | [PLAN-lock.md](./PLAN-lock.md) |
| 10 | `pkg/health` | Health check aggregation | [PLAN-health.md](./PLAN-health.md) |

### Service Infrastructure

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 11 | `pkg/discovery` | Service discovery (Consul, K8s, etcd) | [PLAN-discovery.md](./PLAN-discovery.md) |
| 12 | `pkg/worker` | Worker pool for concurrent tasks | [PLAN-worker.md](./PLAN-worker.md) |
| 13 | `pkg/idgen` | ID generation (UUID, ULID, Snowflake) | [PLAN-idgen.md](./PLAN-idgen.md) |
| 14 | `pkg/feature` | Feature flags and toggles | [PLAN-feature.md](./PLAN-feature.md) |

---

## Communication Packages

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 15 | `pkg/email` | Email with templates (SMTP, SendGrid, SES) | [PLAN-email.md](./PLAN-email.md) |
| 16 | `pkg/websocket` | Real-time WebSocket server | [PLAN-websocket.md](./PLAN-websocket.md) |
| 17 | `pkg/notification` | Push notifications (FCM, APNs) | [PLAN-notification.md](./PLAN-notification.md) |

---

## Observability & Operations

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 18 | `pkg/observability` | Tracing & metrics (OpenTelemetry) | [PLAN-observability.md](./PLAN-observability.md) |
| 19 | `pkg/audit` | Audit logging for compliance | [PLAN-audit.md](./PLAN-audit.md) |

---

## Configuration & Security

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 20 | `pkg/config` | Multi-source configuration | [PLAN-config.md](./PLAN-config.md) |
| 21 | `pkg/secrets` | Secrets management (Vault, AWS, GCP) | [PLAN-secrets.md](./PLAN-secrets.md) |

---

## Data & Validation

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 22 | `pkg/migrate` | Database migrations | [PLAN-migrate.md](./PLAN-migrate.md) |
| 23 | `pkg/validator` | Input validation with i18n | [PLAN-validator.md](./PLAN-validator.md) |
| 24 | `pkg/search` | Search client (Elasticsearch, Meilisearch) | [PLAN-search.md](./PLAN-search.md) |

---

## Application Architecture

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 25 | `pkg/tenant` | Multi-tenancy support | [PLAN-tenant.md](./PLAN-tenant.md) |

---

## Document Generation

| # | Package | Purpose | Plan |
|---|---------|---------|------|
| 26 | `pkg/pdf` | PDF generation | [PLAN-pdf.md](./PLAN-pdf.md) |

---

## Implementation Priority

### Tier 1: Foundation (Implement First)
| Priority | Package | Reason |
|----------|---------|--------|
| 1 | `config` | Foundation for all packages |
| 2 | `validator` | Essential for API security |
| 3 | `migrate` | Database schema management |
| 4 | `health` | Kubernetes deployment ready |
| 5 | `idgen` | Unique ID generation |

### Tier 2: Resilience & Performance
| Priority | Package | Reason |
|----------|---------|--------|
| 6 | `retry` | Resilient operations |
| 7 | `circuitbreaker` | Fault tolerance |
| 8 | `ratelimit` | API protection |
| 9 | `cache` | Performance optimization |
| 10 | `lock` | Distributed coordination |

### Tier 3: Observability & Operations
| Priority | Package | Reason |
|----------|---------|--------|
| 11 | `observability` | Production debugging |
| 12 | `audit` | Compliance & security |
| 13 | `secrets` | Secure credentials |

### Tier 4: Communication & Events
| Priority | Package | Reason |
|----------|---------|--------|
| 14 | `email` | Common requirement |
| 15 | `queue` | Async processing |
| 16 | `pubsub` | Event-driven architecture |
| 17 | `scheduler` | Background jobs |

### Tier 5: Advanced Features
| Priority | Package | Reason |
|----------|---------|--------|
| 18 | `worker` | Controlled concurrency |
| 19 | `feature` | Feature flags |
| 20 | `discovery` | Microservices |
| 21 | `storage` | File handling |
| 22 | `websocket` | Real-time features |
| 23 | `search` | Full-text search |
| 24 | `notification` | Push notifications |
| 25 | `tenant` | Multi-tenancy |
| 26 | `pdf` | Document generation |

---

## Package Independence

Following the existing pattern, each new package will be **completely independent**:

- No imports from other `core-backend` packages
- Can be used standalone or together
- Interfaces allow custom implementations
- All dependencies are optional (pluggable)

---

## Integration Examples

While packages are independent, they work well together:

```go
// Example: Resilient HTTP client
client := httpclient.New(
    httpclient.WithRetry(retry.NewExponentialBackoff(...)),
    httpclient.WithCircuitBreaker(circuitbreaker.New("api", ...)),
    httpclient.WithRateLimit(ratelimit.New(...)),
)

// Example: Configuration with secrets
cfg := config.New(
    config.WithProvider(config.NewEnvProvider()),
    config.WithSecrets(secrets.NewVaultClient()),
)

// Example: Validated API request
func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    if err := validator.Validate(r.Context(), &req); err != nil {
        // Return validation errors
    }
}

// Example: Distributed operations with locking
mutex := lock.NewMutex("resource:123")
mutex.Lock(ctx)
defer mutex.Unlock(ctx)

// Example: Multi-tenant data access
tenant := tenant.FromContext(ctx)
db := strategy.Scope(ctx, tenant.ID)
users, _ := db.Query(ctx, "SELECT * FROM users")

// Example: Async email via queue
queue.Publish("emails", email.NewWelcomeEmail(user))

// Example: Scheduled background jobs
scheduler.Every(1).Hour().Do(cache.DeleteExpired)

// Example: Real-time notifications
ws.BroadcastToRoom(ctx, userID, "notification", data)

// Example: Traced database queries
ctx, span := observability.StartSpan(ctx, "db.query")
rows, _ := postgres.Query(ctx, query)
span.End()
```

---

## Summary

| Category | Count | Packages |
|----------|-------|----------|
| **Core Infrastructure** | 5 | cache, queue, storage, scheduler, pubsub |
| **Resilience** | 5 | ratelimit, circuitbreaker, retry, lock, health |
| **Service Infrastructure** | 4 | discovery, worker, idgen, feature |
| **Communication** | 3 | email, websocket, notification |
| **Observability** | 2 | observability, audit |
| **Configuration** | 2 | config, secrets |
| **Data** | 3 | migrate, validator, search |
| **Architecture** | 1 | tenant |
| **Documents** | 1 | pdf |

**Total: 26 new packages** to complement the existing 4 packages.

---

## Quick Reference

### By Use Case

| Use Case | Packages |
|----------|----------|
| **API Development** | validator, ratelimit, auth |
| **Database** | postgres, migrate, lock |
| **Caching** | cache |
| **Background Processing** | queue, scheduler, worker |
| **Microservices** | discovery, health, circuitbreaker, pubsub |
| **Configuration** | config, secrets, feature |
| **Observability** | observability, audit, health |
| **Communication** | email, websocket, notification |
| **File Handling** | storage, pdf |
| **Multi-tenancy** | tenant |
| **Security** | auth, secrets, ratelimit, audit |
