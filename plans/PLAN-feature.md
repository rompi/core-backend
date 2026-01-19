# Package Plan: pkg/feature

## Overview

A feature flag/toggle package for controlling feature rollouts, A/B testing, and gradual deployments. Supports multiple backends (memory, file, database, LaunchDarkly) with targeting rules, percentage rollouts, and real-time updates.

## Goals

1. **Multiple Backends** - Memory, file, database, LaunchDarkly, Unleash
2. **Targeting Rules** - User segments, attributes, percentages
3. **Flag Types** - Boolean, string, number, JSON
4. **Real-Time Updates** - Live flag changes without restart
5. **A/B Testing** - Variant distribution with tracking
6. **Local Overrides** - Development/testing overrides
7. **Audit Trail** - Track flag evaluations

## Architecture

```
pkg/feature/
├── feature.go            # Core interface
├── config.go             # Configuration
├── options.go            # Functional options
├── flag.go               # Flag definition
├── context.go            # Evaluation context
├── rules.go              # Targeting rules
├── provider/
│   ├── provider.go       # Provider interface
│   ├── memory.go         # In-memory provider
│   ├── file.go           # File-based (JSON/YAML)
│   ├── postgres.go       # PostgreSQL provider
│   ├── redis.go          # Redis provider
│   └── launchdarkly.go   # LaunchDarkly integration
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── examples/
│   ├── basic/
│   ├── percentage-rollout/
│   ├── user-targeting/
│   └── ab-testing/
└── README.md
```

## Core Interfaces

```go
package feature

import (
    "context"
    "time"
)

// Client evaluates feature flags
type Client interface {
    // Bool evaluates a boolean flag
    Bool(ctx context.Context, key string, defaultValue bool) bool

    // String evaluates a string flag
    String(ctx context.Context, key string, defaultValue string) string

    // Int evaluates an integer flag
    Int(ctx context.Context, key string, defaultValue int) int

    // Float evaluates a float flag
    Float(ctx context.Context, key string, defaultValue float64) float64

    // JSON evaluates a JSON flag into target
    JSON(ctx context.Context, key string, target interface{}) error

    // Variation evaluates a flag with full details
    Variation(ctx context.Context, key string) (*Evaluation, error)

    // AllFlags returns all flag values for a context
    AllFlags(ctx context.Context) map[string]interface{}

    // Track records a custom event for analytics
    Track(ctx context.Context, event string, data map[string]interface{})

    // Close releases resources
    Close() error
}

// Flag defines a feature flag
type Flag struct {
    Key           string
    Name          string
    Description   string
    Type          FlagType
    DefaultValue  interface{}
    Enabled       bool
    Rules         []Rule
    Variants      []Variant
    Prerequisites []Prerequisite
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// FlagType defines the flag value type
type FlagType string

const (
    FlagTypeBool   FlagType = "bool"
    FlagTypeString FlagType = "string"
    FlagTypeInt    FlagType = "int"
    FlagTypeFloat  FlagType = "float"
    FlagTypeJSON   FlagType = "json"
)

// Rule defines targeting rules
type Rule struct {
    ID         string
    Clauses    []Clause
    Variation  int
    Percentage *int
    Rollout    *Rollout
}

// Clause defines a condition
type Clause struct {
    Attribute string
    Operator  Operator
    Values    []interface{}
    Negate    bool
}

// Operator for clause evaluation
type Operator string

const (
    OpEquals      Operator = "eq"
    OpNotEquals   Operator = "neq"
    OpContains    Operator = "contains"
    OpStartsWith  Operator = "startsWith"
    OpEndsWith    Operator = "endsWith"
    OpIn          Operator = "in"
    OpNotIn       Operator = "notIn"
    OpGreaterThan Operator = "gt"
    OpLessThan    Operator = "lt"
    OpMatches     Operator = "matches" // Regex
    OpSemVer      Operator = "semver"
)

// Variant defines a flag variation
type Variant struct {
    Name   string
    Value  interface{}
    Weight int
}

// Rollout defines percentage-based rollout
type Rollout struct {
    BucketBy   string // Attribute for consistent bucketing
    Variations []WeightedVariation
}

// WeightedVariation for percentage distribution
type WeightedVariation struct {
    Variation int
    Weight    int // Out of 100000 (0.001% precision)
}

// Evaluation contains flag evaluation result
type Evaluation struct {
    Key           string
    Value         interface{}
    VariationIdx  int
    Reason        EvaluationReason
    RuleID        string
    InExperiment  bool
    PrerequisiteFailed bool
}

// EvaluationReason explains why a value was returned
type EvaluationReason string

const (
    ReasonFallthrough   EvaluationReason = "FALLTHROUGH"
    ReasonTargetMatch   EvaluationReason = "TARGET_MATCH"
    ReasonRuleMatch     EvaluationReason = "RULE_MATCH"
    ReasonPrerequisite  EvaluationReason = "PREREQUISITE_FAILED"
    ReasonOff           EvaluationReason = "OFF"
    ReasonError         EvaluationReason = "ERROR"
)
```

## Evaluation Context

```go
// Context holds attributes for flag evaluation
type Context struct {
    // Key is the unique user/entity identifier
    Key string

    // Name is the display name
    Name string

    // Email is the email address
    Email string

    // IP is the IP address
    IP string

    // Country is the country code
    Country string

    // Custom holds custom attributes
    Custom map[string]interface{}

    // Anonymous indicates an anonymous user
    Anonymous bool

    // Groups holds group memberships
    Groups []string
}

// NewContext creates an evaluation context
func NewContext(key string) *Context

// WithAttribute adds a custom attribute
func (c *Context) WithAttribute(key string, value interface{}) *Context

// WithGroup adds group membership
func (c *Context) WithGroup(group string) *Context

// ContextFromHTTPRequest extracts context from request
func ContextFromHTTPRequest(r *http.Request) *Context
```

## Configuration

```go
// Config holds feature flag configuration
type Config struct {
    // Provider: "memory", "file", "postgres", "redis", "launchdarkly"
    Provider string `env:"FEATURE_PROVIDER" default:"memory"`

    // Refresh interval for remote providers
    RefreshInterval time.Duration `env:"FEATURE_REFRESH_INTERVAL" default:"30s"`

    // Enable offline mode
    OfflineMode bool `env:"FEATURE_OFFLINE" default:"false"`

    // Send analytics events
    SendEvents bool `env:"FEATURE_SEND_EVENTS" default:"true"`
}

// FileConfig for file-based provider
type FileConfig struct {
    Path       string `env:"FEATURE_FILE_PATH" default:"features.yaml"`
    WatchFile  bool   `env:"FEATURE_FILE_WATCH" default:"true"`
}

// PostgresConfig for database provider
type PostgresConfig struct {
    TableName string `env:"FEATURE_TABLE" default:"feature_flags"`
}

// LaunchDarklyConfig for LaunchDarkly
type LaunchDarklyConfig struct {
    SDKKey     string `env:"LAUNCHDARKLY_SDK_KEY" required:"true"`
    BaseURL    string `env:"LAUNCHDARKLY_BASE_URL"`
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/feature"
)

func main() {
    // Create client
    client, err := feature.New(feature.Config{
        Provider: "file",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Evaluate boolean flag
    if client.Bool(ctx, "new-dashboard", false) {
        showNewDashboard()
    } else {
        showOldDashboard()
    }

    // Evaluate string flag
    theme := client.String(ctx, "theme", "light")
    applyTheme(theme)
}
```

### With User Context

```go
func main() {
    client, _ := feature.New(cfg)

    // Create user context
    userCtx := feature.NewContext("user-123").
        WithAttribute("email", "john@example.com").
        WithAttribute("plan", "pro").
        WithAttribute("country", "US").
        WithGroup("beta-testers")

    ctx := feature.WithContext(context.Background(), userCtx)

    // Flag evaluated based on user attributes
    if client.Bool(ctx, "advanced-features", false) {
        showAdvancedFeatures()
    }
}
```

### Percentage Rollout

```yaml
# features.yaml
flags:
  new-checkout:
    type: bool
    default: false
    enabled: true
    rules:
      - rollout:
          bucketBy: key  # Consistent bucketing by user key
          variations:
            - variation: 0  # false
              weight: 80000  # 80%
            - variation: 1  # true
              weight: 20000  # 20%
    variants:
      - name: off
        value: false
      - name: on
        value: true
```

```go
func main() {
    client, _ := feature.New(cfg)

    userCtx := feature.NewContext("user-123")
    ctx := feature.WithContext(context.Background(), userCtx)

    // 20% of users get new checkout
    // Same user always gets same result (consistent bucketing)
    if client.Bool(ctx, "new-checkout", false) {
        showNewCheckout()
    }
}
```

### Targeting Rules

```yaml
# features.yaml
flags:
  premium-features:
    type: bool
    default: false
    enabled: true
    rules:
      # Rule 1: Beta testers get it
      - clauses:
          - attribute: groups
            operator: contains
            values: ["beta-testers"]
        variation: 1

      # Rule 2: Pro plan users get it
      - clauses:
          - attribute: plan
            operator: in
            values: ["pro", "enterprise"]
        variation: 1

      # Rule 3: US users get 50% rollout
      - clauses:
          - attribute: country
            operator: eq
            values: ["US"]
        rollout:
          bucketBy: key
          variations:
            - variation: 0
              weight: 50000
            - variation: 1
              weight: 50000
```

### A/B Testing

```go
func main() {
    client, _ := feature.New(cfg)

    userCtx := feature.NewContext("user-123")
    ctx := feature.WithContext(context.Background(), userCtx)

    // Get variant with full details
    eval, _ := client.Variation(ctx, "checkout-button-color")

    buttonColor := eval.Value.(string) // "blue", "green", or "red"

    // Track for analytics
    if eval.InExperiment {
        client.Track(ctx, "checkout-button-view", map[string]interface{}{
            "variant": eval.VariationIdx,
        })
    }

    renderButton(buttonColor)
}
```

### Local Overrides (Development)

```go
func main() {
    client, _ := feature.New(cfg,
        feature.WithOverrides(map[string]interface{}{
            "new-feature":     true,
            "experiment-mode": "variant-b",
        }),
    )

    // Overrides take precedence during development
    client.Bool(ctx, "new-feature", false) // Always true
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/feature/middleware"
)

func main() {
    client, _ := feature.New(cfg)

    // Middleware extracts user context from request
    mw := middleware.HTTP(client,
        middleware.WithContextExtractor(func(r *http.Request) *feature.Context {
            user := getUserFromRequest(r)
            return feature.NewContext(user.ID).
                WithAttribute("plan", user.Plan)
        }),
    )

    mux := http.NewServeMux()
    http.ListenAndServe(":8080", mw(mux))
}

// In handlers, use context from request
func handler(w http.ResponseWriter, r *http.Request) {
    if feature.Bool(r.Context(), "new-feature", false) {
        // New feature enabled for this user
    }
}
```

### Prerequisites

```yaml
flags:
  advanced-analytics:
    type: bool
    default: false
    enabled: true
    prerequisites:
      - key: basic-analytics  # Must be enabled first
        variation: 1
    rules:
      - variation: 1  # Enable for all if prerequisite met
```

## Dependencies

- **Required:** None (memory/file providers)
- **Optional:**
  - `github.com/launchdarkly/go-server-sdk` for LaunchDarkly
  - Database drivers for database provider

## Implementation Phases

### Phase 1: Core Interface & Memory Provider
1. Define Client, Flag interfaces
2. Memory provider
3. Basic boolean evaluation

### Phase 2: File Provider
1. YAML/JSON file provider
2. File watching
3. All flag types

### Phase 3: Targeting Rules
1. Clause evaluation
2. Rule matching
3. Operators (eq, contains, in, etc.)

### Phase 4: Rollouts
1. Percentage rollout
2. Consistent bucketing
3. Weighted variants

### Phase 5: Database Provider
1. PostgreSQL provider
2. Redis provider (for caching)
3. Real-time updates

### Phase 6: Advanced Features
1. Prerequisites
2. A/B testing support
3. Analytics tracking
4. LaunchDarkly integration

### Phase 7: Documentation
1. README
2. Examples
3. Best practices

### Phase 8: Performance Metrics & Monitoring
1. Evaluation latency metrics
2. Cache hit/miss tracking
3. Error rate monitoring
4. Provider health checks
5. Dashboard integration

## Performance Metrics

### Metrics Interface

```go
// Metrics defines the interface for collecting feature flag metrics
type Metrics interface {
    // RecordEvaluation records a flag evaluation
    RecordEvaluation(key string, duration time.Duration, reason EvaluationReason)

    // RecordError records an evaluation error
    RecordError(key string, err error)

    // RecordCacheHit records a cache hit
    RecordCacheHit(key string)

    // RecordCacheMiss records a cache miss
    RecordCacheMiss(key string)

    // RecordProviderLatency records provider response time
    RecordProviderLatency(provider string, operation string, duration time.Duration)

    // RecordProviderError records provider errors
    RecordProviderError(provider string, operation string, err error)
}

// MetricsConfig configures metrics collection
type MetricsConfig struct {
    // Enabled enables metrics collection
    Enabled bool `env:"FEATURE_METRICS_ENABLED" default:"true"`

    // Namespace is the metrics namespace prefix
    Namespace string `env:"FEATURE_METRICS_NAMESPACE" default:"feature_flags"`

    // HistogramBuckets for latency histograms (in seconds)
    HistogramBuckets []float64 `default:"0.001,0.005,0.01,0.025,0.05,0.1,0.25,0.5,1"`

    // EnablePerFlagMetrics enables per-flag dimension (can be high cardinality)
    EnablePerFlagMetrics bool `env:"FEATURE_METRICS_PER_FLAG" default:"false"`
}
```

### Prometheus Metrics Implementation

```go
// PrometheusMetrics implements Metrics using Prometheus
type PrometheusMetrics struct {
    evaluationDuration  *prometheus.HistogramVec
    evaluationTotal     *prometheus.CounterVec
    evaluationErrors    *prometheus.CounterVec
    cacheHits           *prometheus.CounterVec
    cacheMisses         *prometheus.CounterVec
    providerLatency     *prometheus.HistogramVec
    providerErrors      *prometheus.CounterVec
    flagsTotal          prometheus.Gauge
    activeFlags         prometheus.Gauge
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(cfg MetricsConfig) *PrometheusMetrics

// Metric names:
// - feature_flags_evaluation_duration_seconds (histogram)
// - feature_flags_evaluation_total (counter)
// - feature_flags_evaluation_errors_total (counter)
// - feature_flags_cache_hits_total (counter)
// - feature_flags_cache_misses_total (counter)
// - feature_flags_provider_latency_seconds (histogram)
// - feature_flags_provider_errors_total (counter)
// - feature_flags_total (gauge)
// - feature_flags_active_total (gauge)
```

### Key Metrics to Track

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `evaluation_duration_seconds` | Histogram | `flag_key`, `reason` | Time to evaluate a flag |
| `evaluation_total` | Counter | `flag_key`, `reason`, `value` | Total evaluations |
| `evaluation_errors_total` | Counter | `flag_key`, `error_type` | Evaluation errors |
| `cache_hits_total` | Counter | `flag_key` | Cache hits |
| `cache_misses_total` | Counter | `flag_key` | Cache misses |
| `provider_latency_seconds` | Histogram | `provider`, `operation` | Provider response time |
| `provider_errors_total` | Counter | `provider`, `operation`, `error_type` | Provider errors |
| `flags_total` | Gauge | - | Total number of flags |
| `flags_active_total` | Gauge | - | Number of enabled flags |

### Usage with Metrics

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/user/core-backend/pkg/feature"
)

func main() {
    // Create Prometheus metrics
    metrics := feature.NewPrometheusMetrics(feature.MetricsConfig{
        Enabled:              true,
        Namespace:            "myapp_feature_flags",
        EnablePerFlagMetrics: true,
    })

    // Register with Prometheus
    prometheus.MustRegister(metrics)

    // Create client with metrics
    client, err := feature.New(cfg,
        feature.WithMetrics(metrics),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Evaluations automatically record metrics
    client.Bool(ctx, "new-feature", false)
}
```

### OpenTelemetry Integration

```go
// OTelMetrics implements Metrics using OpenTelemetry
type OTelMetrics struct {
    meter           metric.Meter
    evalDuration    metric.Float64Histogram
    evalCounter     metric.Int64Counter
    errorCounter    metric.Int64Counter
    cacheHitCounter metric.Int64Counter
}

// NewOTelMetrics creates an OpenTelemetry metrics collector
func NewOTelMetrics(meter metric.Meter, cfg MetricsConfig) (*OTelMetrics, error)
```

### Health Check Endpoint

```go
// HealthCheck returns the health status of the feature flag system
type HealthCheck struct {
    Status           string            `json:"status"`  // "healthy", "degraded", "unhealthy"
    Provider         ProviderHealth    `json:"provider"`
    Cache            CacheHealth       `json:"cache"`
    LastRefresh      time.Time         `json:"last_refresh"`
    FlagsLoaded      int               `json:"flags_loaded"`
    EvaluationStats  EvaluationStats   `json:"evaluation_stats"`
}

type ProviderHealth struct {
    Connected    bool          `json:"connected"`
    Latency      time.Duration `json:"latency_ms"`
    LastError    string        `json:"last_error,omitempty"`
    LastErrorAt  *time.Time    `json:"last_error_at,omitempty"`
}

type CacheHealth struct {
    Enabled   bool    `json:"enabled"`
    HitRate   float64 `json:"hit_rate"`
    Size      int     `json:"size"`
    MaxSize   int     `json:"max_size"`
}

type EvaluationStats struct {
    TotalEvaluations int64   `json:"total_evaluations"`
    ErrorRate        float64 `json:"error_rate"`
    AvgLatencyMs     float64 `json:"avg_latency_ms"`
    P99LatencyMs     float64 `json:"p99_latency_ms"`
}

// HTTP handler for health checks
func HealthHandler(client Client) http.HandlerFunc
```

### Dashboard Recommendations

#### Grafana Dashboard Panels

1. **Evaluation Performance**
   - Evaluation latency (p50, p95, p99)
   - Evaluations per second by flag
   - Error rate trend

2. **Cache Efficiency**
   - Cache hit ratio over time
   - Cache size vs capacity
   - Miss rate by flag

3. **Provider Health**
   - Provider latency histogram
   - Provider error rate
   - Connection status

4. **Flag Activity**
   - Active vs inactive flags
   - Most evaluated flags (top 10)
   - Evaluation breakdown by reason

#### Sample Prometheus Queries

```promql
# Evaluation latency P99
histogram_quantile(0.99,
  sum(rate(feature_flags_evaluation_duration_seconds_bucket[5m])) by (le)
)

# Error rate
sum(rate(feature_flags_evaluation_errors_total[5m]))
/ sum(rate(feature_flags_evaluation_total[5m])) * 100

# Cache hit ratio
sum(rate(feature_flags_cache_hits_total[5m]))
/ (sum(rate(feature_flags_cache_hits_total[5m])) + sum(rate(feature_flags_cache_misses_total[5m]))) * 100

# Provider latency P95
histogram_quantile(0.95,
  sum(rate(feature_flags_provider_latency_seconds_bucket[5m])) by (le, provider)
)

# Top 10 most evaluated flags
topk(10, sum(rate(feature_flags_evaluation_total[1h])) by (flag_key))
```

### Alerting Rules

```yaml
# Prometheus alerting rules
groups:
  - name: feature_flags
    rules:
      - alert: FeatureFlagHighErrorRate
        expr: |
          sum(rate(feature_flags_evaluation_errors_total[5m]))
          / sum(rate(feature_flags_evaluation_total[5m])) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Feature flag error rate above 1%"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: FeatureFlagHighLatency
        expr: |
          histogram_quantile(0.99,
            sum(rate(feature_flags_evaluation_duration_seconds_bucket[5m])) by (le)
          ) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Feature flag P99 latency above 100ms"
          description: "P99 latency is {{ $value | humanizeDuration }}"

      - alert: FeatureFlagProviderDown
        expr: feature_flags_provider_errors_total > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Feature flag provider experiencing errors"

      - alert: FeatureFlagLowCacheHitRate
        expr: |
          sum(rate(feature_flags_cache_hits_total[5m]))
          / (sum(rate(feature_flags_cache_hits_total[5m])) + sum(rate(feature_flags_cache_misses_total[5m]))) < 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Feature flag cache hit rate below 80%"
```

### Event Tracking for Analytics

```go
// AnalyticsEvent for tracking flag evaluations
type AnalyticsEvent struct {
    Timestamp    time.Time              `json:"timestamp"`
    FlagKey      string                 `json:"flag_key"`
    UserKey      string                 `json:"user_key"`
    Value        interface{}            `json:"value"`
    VariationIdx int                    `json:"variation_idx"`
    Reason       EvaluationReason       `json:"reason"`
    RuleID       string                 `json:"rule_id,omitempty"`
    InExperiment bool                   `json:"in_experiment"`
    Context      map[string]interface{} `json:"context,omitempty"`
}

// AnalyticsExporter exports evaluation events
type AnalyticsExporter interface {
    Export(ctx context.Context, events []AnalyticsEvent) error
    Flush(ctx context.Context) error
    Close() error
}

// Built-in exporters:
// - StdoutExporter (for debugging)
// - FileExporter (JSON lines)
// - HTTPExporter (webhook)
// - KafkaExporter (for streaming analytics)
```
