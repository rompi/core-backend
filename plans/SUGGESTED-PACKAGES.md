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

### 4. pkg/scheduler - Background Job Processing

**Purpose:** Background job scheduling and processing with cron support.

**Why:** Applications need to run periodic tasks (cleanup, reports, sync) and process async jobs.

**Plan:** [PLAN-scheduler.md](./PLAN-scheduler.md)

---

## Communication Packages

### 5. pkg/email - Email Service

**Purpose:** Email sending with template support, multiple providers (SMTP, SendGrid, SES).

**Why:** Nearly every application needs to send transactional emails (welcome, password reset, notifications).

**Plan:** [PLAN-email.md](./PLAN-email.md)

---

### 6. pkg/websocket - Real-Time Communication

**Purpose:** WebSocket server with rooms/channels, pub/sub, and horizontal scaling via Redis.

**Why:** Real-time features (chat, notifications, live updates) require persistent connections.

**Plan:** [PLAN-websocket.md](./PLAN-websocket.md)

---

### 7. pkg/notification - Push Notifications

**Purpose:** Push notifications via FCM, APNs, and Web Push.

**Why:** Mobile and web apps need push notifications for user engagement.

**Plan:** [PLAN-notification.md](./PLAN-notification.md)

---

## Observability & Operations

### 8. pkg/observability - Tracing & Metrics

**Purpose:** Distributed tracing and metrics collection using OpenTelemetry.

**Why:** Production systems require observability for debugging, performance monitoring, and alerting.

**Plan:** [PLAN-observability.md](./PLAN-observability.md)

---

## Configuration & Security

### 9. pkg/config - Configuration Management

**Purpose:** Unified configuration from multiple sources (env, files, Consul, etcd).

**Why:** Applications need flexible configuration that works across environments.

**Plan:** [PLAN-config.md](./PLAN-config.md)

---

### 10. pkg/secrets - Secrets Management

**Purpose:** Secure secrets retrieval from Vault, AWS Secrets Manager, GCP, Azure.

**Why:** Production apps need secure secrets management beyond environment variables.

**Plan:** [PLAN-secrets.md](./PLAN-secrets.md)

---

## Data & Validation

### 11. pkg/migrate - Database Migrations

**Purpose:** Schema versioning with up/down migrations for PostgreSQL, MySQL, SQLite.

**Why:** Every database-backed application needs schema migration management.

**Plan:** [PLAN-migrate.md](./PLAN-migrate.md)

---

### 12. pkg/validator - Input Validation

**Purpose:** Struct tag-based validation with custom rules and i18n error messages.

**Why:** Input validation is critical for security and data integrity.

**Plan:** [PLAN-validator.md](./PLAN-validator.md)

---

### 13. pkg/search - Search Engine Client

**Purpose:** Unified search client for Elasticsearch, Meilisearch, Typesense.

**Why:** Applications with search functionality need a consistent search interface.

**Plan:** [PLAN-search.md](./PLAN-search.md)

---

## Document Generation

### 14. pkg/pdf - PDF Generation

**Purpose:** Create PDFs from HTML, templates, or programmatically.

**Why:** Applications need to generate reports, invoices, and documents.

**Plan:** [PLAN-pdf.md](./PLAN-pdf.md)

---

## Implementation Priority

Recommended implementation order based on common usage patterns:

| Priority | Package | Reason |
|----------|---------|--------|
| 1 | `config` | Foundation for all other packages |
| 2 | `validator` | Essential for API input validation |
| 3 | `migrate` | Required for database schema management |
| 4 | `cache` | Most applications need caching |
| 5 | `observability` | Critical for production debugging |
| 6 | `secrets` | Secure credential management |
| 7 | `email` | Common requirement for auth flows |
| 8 | `queue` | Essential for async processing |
| 9 | `scheduler` | Background job processing |
| 10 | `storage` | File handling needs |
| 11 | `websocket` | Real-time features |
| 12 | `search` | Full-text search functionality |
| 13 | `notification` | Push notification support |
| 14 | `pdf` | Document generation |

## Package Independence

Following the existing pattern, each new package will be **completely independent**:

- No imports from other `core-backend` packages
- Can be used standalone or together
- Interfaces allow custom implementations
- All dependencies are optional (pluggable)

## Integration Examples

While packages are independent, they work well together:

```go
// Example: Load configuration with secrets
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

// Example: Real-time notifications
ws.BroadcastToRoom(ctx, userID, "notification", notification)
```

## Summary

| Category | Packages |
|----------|----------|
| **Infrastructure** | cache, queue, storage, scheduler |
| **Communication** | email, websocket, notification |
| **Observability** | observability |
| **Configuration** | config, secrets |
| **Data** | migrate, validator, search |
| **Documents** | pdf |

**Total: 14 new packages** to complement the existing 4 packages.
