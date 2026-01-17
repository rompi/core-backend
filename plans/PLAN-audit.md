# Package Plan: pkg/audit

## Overview

An audit logging package for tracking system activities, user actions, and data changes. Provides structured audit trails for compliance (SOC2, HIPAA, GDPR), security analysis, and debugging with support for multiple storage backends.

## Goals

1. **Structured Logging** - Consistent audit event format
2. **Multiple Backends** - Database, file, Elasticsearch, cloud services
3. **Automatic Tracking** - Middleware for HTTP/gRPC
4. **Data Change Tracking** - Before/after snapshots
5. **Query Interface** - Search and filter audit logs
6. **Retention Policies** - Automatic cleanup
7. **Compliance Ready** - SOC2, HIPAA, GDPR support

## Architecture

```
pkg/audit/
├── audit.go              # Core Auditor interface
├── config.go             # Configuration
├── options.go            # Functional options
├── event.go              # Audit event structure
├── context.go            # Context extraction
├── store/
│   ├── store.go          # Store interface
│   ├── postgres.go       # PostgreSQL store
│   ├── elasticsearch.go  # Elasticsearch store
│   ├── file.go           # File-based store
│   └── memory.go         # In-memory (testing)
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── diff/
│   ├── diff.go           # Change detection
│   └── json.go           # JSON diff
├── examples/
│   ├── basic/
│   ├── http-middleware/
│   └── data-changes/
└── README.md
```

## Core Interfaces

```go
package audit

import (
    "context"
    "time"
)

// Auditor records audit events
type Auditor interface {
    // Log records an audit event
    Log(ctx context.Context, event *Event) error

    // LogAsync records asynchronously
    LogAsync(ctx context.Context, event *Event)

    // Query searches audit logs
    Query(ctx context.Context, filter Filter) (*QueryResult, error)

    // Close releases resources
    Close() error
}

// Event represents an audit log entry
type Event struct {
    // ID is the unique event identifier
    ID string

    // Timestamp when the event occurred
    Timestamp time.Time

    // Actor who performed the action
    Actor Actor

    // Action performed
    Action string

    // Resource affected
    Resource Resource

    // Outcome of the action
    Outcome Outcome

    // Request details
    Request *RequestInfo

    // Changes made (for data modifications)
    Changes *Changes

    // Metadata for additional context
    Metadata map[string]interface{}

    // Tags for categorization
    Tags []string
}

// Actor represents who performed the action
type Actor struct {
    // ID is the actor identifier
    ID string

    // Type: "user", "service", "system", "api_key"
    Type string

    // Name is the display name
    Name string

    // Email address
    Email string

    // IP address
    IP string

    // Session ID
    SessionID string

    // Additional attributes
    Attributes map[string]string
}

// Resource represents what was affected
type Resource struct {
    // Type: "user", "order", "payment", etc.
    Type string

    // ID is the resource identifier
    ID string

    // Name is the resource name
    Name string

    // Parent resource (optional)
    Parent *Resource
}

// Outcome of the action
type Outcome string

const (
    OutcomeSuccess Outcome = "success"
    OutcomeFailure Outcome = "failure"
    OutcomeDenied  Outcome = "denied"
    OutcomePending Outcome = "pending"
)

// RequestInfo contains request details
type RequestInfo struct {
    Method    string
    Path      string
    Query     map[string]string
    Headers   map[string]string
    Body      []byte
    UserAgent string
    Referer   string
}

// Changes represents data modifications
type Changes struct {
    Before interface{}
    After  interface{}
    Diff   []FieldChange
}

// FieldChange represents a single field change
type FieldChange struct {
    Field    string
    Before   interface{}
    After    interface{}
    Action   string // "added", "modified", "removed"
}

// Filter for querying audit logs
type Filter struct {
    // Time range
    StartTime time.Time
    EndTime   time.Time

    // Actor filters
    ActorID   string
    ActorType string

    // Resource filters
    ResourceType string
    ResourceID   string

    // Action filter
    Action  string
    Actions []string

    // Outcome filter
    Outcome Outcome

    // Full-text search
    Query string

    // Pagination
    Limit  int
    Offset int

    // Sort order
    OrderBy   string
    OrderDesc bool
}

// QueryResult contains query results
type QueryResult struct {
    Events     []*Event
    Total      int64
    HasMore    bool
    NextOffset int
}
```

## Configuration

```go
// Config holds audit configuration
type Config struct {
    // Store type: "postgres", "elasticsearch", "file", "memory"
    Store string `env:"AUDIT_STORE" default:"postgres"`

    // Async mode for non-blocking logging
    Async bool `env:"AUDIT_ASYNC" default:"true"`

    // Buffer size for async mode
    BufferSize int `env:"AUDIT_BUFFER_SIZE" default:"1000"`

    // Flush interval for async mode
    FlushInterval time.Duration `env:"AUDIT_FLUSH_INTERVAL" default:"5s"`

    // Retention period
    Retention time.Duration `env:"AUDIT_RETENTION" default:"2160h"` // 90 days

    // Fields to redact from logs
    RedactFields []string `env:"AUDIT_REDACT_FIELDS"`

    // Include request body
    IncludeRequestBody bool `env:"AUDIT_INCLUDE_BODY" default:"false"`

    // Maximum body size
    MaxBodySize int `env:"AUDIT_MAX_BODY_SIZE" default:"10240"`
}

// PostgresConfig for PostgreSQL store
type PostgresConfig struct {
    TableName string `env:"AUDIT_TABLE" default:"audit_logs"`
    Schema    string `env:"AUDIT_SCHEMA" default:"public"`
}

// ElasticsearchConfig for Elasticsearch store
type ElasticsearchConfig struct {
    URLs      []string `env:"AUDIT_ES_URLS" default:"http://localhost:9200"`
    Index     string   `env:"AUDIT_ES_INDEX" default:"audit-logs"`
    Username  string   `env:"AUDIT_ES_USERNAME"`
    Password  string   `env:"AUDIT_ES_PASSWORD"`
}
```

## Usage Examples

### Basic Audit Logging

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/audit"
)

func main() {
    auditor, err := audit.New(audit.Config{
        Store: "postgres",
        Async: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer auditor.Close()

    ctx := context.Background()

    // Log an event
    auditor.Log(ctx, &audit.Event{
        Action: "user.login",
        Actor: audit.Actor{
            ID:    "user-123",
            Type:  "user",
            Name:  "John Doe",
            Email: "john@example.com",
            IP:    "192.168.1.100",
        },
        Resource: audit.Resource{
            Type: "session",
            ID:   "sess-456",
        },
        Outcome: audit.OutcomeSuccess,
        Metadata: map[string]interface{}{
            "login_method": "password",
            "mfa_used":     true,
        },
    })
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/audit"
    "github.com/user/core-backend/pkg/audit/middleware"
)

func main() {
    auditor, _ := audit.New(cfg)

    // Create middleware
    auditMiddleware := middleware.HTTP(auditor,
        middleware.WithActorExtractor(func(r *http.Request) audit.Actor {
            user := getUserFromContext(r.Context())
            return audit.Actor{
                ID:    user.ID,
                Type:  "user",
                Name:  user.Name,
                Email: user.Email,
                IP:    r.RemoteAddr,
            }
        }),
        middleware.WithActionMapper(func(r *http.Request) string {
            return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
        }),
        middleware.WithSkipPaths("/health", "/metrics"),
    )

    mux := http.NewServeMux()
    http.ListenAndServe(":8080", auditMiddleware(mux))
}
```

### Data Change Tracking

```go
func updateUser(ctx context.Context, userID string, updates UserUpdate) error {
    // Get before state
    before, _ := db.GetUser(ctx, userID)

    // Apply updates
    after, err := db.UpdateUser(ctx, userID, updates)
    if err != nil {
        auditor.Log(ctx, &audit.Event{
            Action:   "user.update",
            Resource: audit.Resource{Type: "user", ID: userID},
            Outcome:  audit.OutcomeFailure,
            Metadata: map[string]interface{}{"error": err.Error()},
        })
        return err
    }

    // Log with changes
    auditor.Log(ctx, &audit.Event{
        Action:   "user.update",
        Actor:    getActorFromContext(ctx),
        Resource: audit.Resource{Type: "user", ID: userID},
        Outcome:  audit.OutcomeSuccess,
        Changes:  audit.Diff(before, after),
    })

    return nil
}
```

### Querying Audit Logs

```go
func main() {
    auditor, _ := audit.New(cfg)

    ctx := context.Background()

    // Query recent login failures
    result, err := auditor.Query(ctx, audit.Filter{
        StartTime:    time.Now().Add(-24 * time.Hour),
        EndTime:      time.Now(),
        Action:       "user.login",
        Outcome:      audit.OutcomeFailure,
        Limit:        100,
        OrderBy:      "timestamp",
        OrderDesc:    true,
    })

    for _, event := range result.Events {
        fmt.Printf("%s: %s from %s\n",
            event.Timestamp,
            event.Actor.Email,
            event.Actor.IP,
        )
    }
}
```

### Security Sensitive Actions

```go
// Define security-sensitive actions
var sensitiveActions = []string{
    "user.password_change",
    "user.mfa_disable",
    "role.permission_grant",
    "api_key.create",
    "data.export",
}

func logSensitiveAction(ctx context.Context, action string, resource audit.Resource) {
    auditor.Log(ctx, &audit.Event{
        Action:   action,
        Actor:    getActorFromContext(ctx),
        Resource: resource,
        Outcome:  audit.OutcomeSuccess,
        Tags:     []string{"security", "sensitive"},
        Metadata: map[string]interface{}{
            "security_level": "high",
        },
    })

    // Also send alert for critical actions
    if isCriticalAction(action) {
        alerting.Send("Sensitive action: " + action)
    }
}
```

### Compliance Export

```go
func exportAuditLogs(ctx context.Context, startDate, endDate time.Time) ([]byte, error) {
    auditor, _ := audit.New(cfg)

    result, err := auditor.Query(ctx, audit.Filter{
        StartTime: startDate,
        EndTime:   endDate,
        Limit:     10000,
    })
    if err != nil {
        return nil, err
    }

    // Export to CSV/JSON for compliance
    return audit.ExportCSV(result.Events)
}
```

### Redacting Sensitive Data

```go
func main() {
    auditor, _ := audit.New(audit.Config{
        RedactFields: []string{
            "password",
            "ssn",
            "credit_card",
            "api_key",
        },
    })

    // Sensitive fields are automatically redacted
    auditor.Log(ctx, &audit.Event{
        Action: "user.create",
        Changes: &audit.Changes{
            After: map[string]interface{}{
                "email":    "user@example.com",
                "password": "secret123", // Will be logged as "***REDACTED***"
            },
        },
    })
}
```

## Database Schema (PostgreSQL)

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_id VARCHAR(255),
    actor_type VARCHAR(50),
    actor_name VARCHAR(255),
    actor_email VARCHAR(255),
    actor_ip INET,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    outcome VARCHAR(50) NOT NULL,
    changes JSONB,
    metadata JSONB,
    tags TEXT[],
    request_method VARCHAR(10),
    request_path TEXT,
    request_body BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_outcome ON audit_logs(outcome);
```

## Dependencies

- **Required:** None (memory store)
- **Optional:**
  - `github.com/jackc/pgx/v5` for PostgreSQL
  - `github.com/elastic/go-elasticsearch` for Elasticsearch

## Implementation Phases

### Phase 1: Core Interface
1. Define Auditor, Event interfaces
2. Memory store
3. Basic logging

### Phase 2: PostgreSQL Store
1. Schema design
2. Query implementation
3. Retention cleanup

### Phase 3: Middleware
1. HTTP middleware
2. gRPC interceptor
3. Context extraction

### Phase 4: Change Tracking
1. Diff algorithm
2. JSON diff
3. Field redaction

### Phase 5: Advanced Features
1. Elasticsearch store
2. Export capabilities
3. Compliance helpers

### Phase 6: Documentation
1. README
2. Examples
3. Schema documentation
