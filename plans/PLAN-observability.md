# Package Plan: pkg/observability

## Overview

A comprehensive observability package providing distributed tracing, metrics collection, and structured logging integration. Built on OpenTelemetry for vendor-neutral instrumentation with support for popular backends (Jaeger, Zipkin, Prometheus, OTLP).

## Goals

1. **Distributed Tracing** - Trace requests across services
2. **Metrics Collection** - Counters, gauges, histograms
3. **Context Propagation** - Automatic trace context passing
4. **Multiple Exporters** - Jaeger, Zipkin, Prometheus, OTLP
5. **Middleware Integration** - HTTP and gRPC instrumentation
6. **Structured Logging** - Trace-correlated log entries
7. **Minimal Overhead** - Sampling and batching for performance

## Architecture

```
pkg/observability/
├── observability.go      # Core types and initialization
├── config.go             # Configuration with env support
├── options.go            # Functional options
├── tracer.go             # Tracing functionality
├── span.go               # Span creation and management
├── metrics.go            # Metrics collection
├── meter.go              # Meter provider setup
├── propagation.go        # Context propagation
├── middleware/
│   ├── http.go           # HTTP middleware
│   ├── grpc.go           # gRPC interceptors
│   └── httpclient.go     # HTTP client middleware
├── exporters/
│   ├── jaeger.go         # Jaeger exporter
│   ├── zipkin.go         # Zipkin exporter
│   ├── prometheus.go     # Prometheus exporter
│   ├── otlp.go           # OTLP exporter
│   └── stdout.go         # Stdout for development
├── logging/
│   ├── adapter.go        # Logger adapters
│   └── fields.go         # Trace-correlated fields
├── examples/
│   ├── basic/
│   ├── http-server/
│   ├── grpc-server/
│   └── microservices/
└── README.md
```

## Core Interfaces

```go
package observability

import (
    "context"
    "time"
)

// Provider manages observability resources
type Provider interface {
    // Tracer returns a tracer for creating spans
    Tracer(name string, opts ...TracerOption) Tracer

    // Meter returns a meter for recording metrics
    Meter(name string, opts ...MeterOption) Meter

    // Shutdown gracefully shuts down all exporters
    Shutdown(ctx context.Context) error
}

// Tracer creates spans for distributed tracing
type Tracer interface {
    // Start creates a new span
    Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a unit of work
type Span interface {
    // End completes the span
    End(opts ...SpanEndOption)

    // SetAttributes adds attributes to the span
    SetAttributes(attrs ...Attribute)

    // SetStatus sets the span status
    SetStatus(code StatusCode, description string)

    // RecordError records an error
    RecordError(err error, opts ...EventOption)

    // AddEvent adds an event to the span
    AddEvent(name string, opts ...EventOption)

    // SpanContext returns the span's context
    SpanContext() SpanContext

    // IsRecording returns true if span is recording
    IsRecording() bool
}

// Meter records metrics
type Meter interface {
    // Counter creates a counter instrument
    Counter(name string, opts ...InstrumentOption) Counter

    // UpDownCounter creates an up-down counter
    UpDownCounter(name string, opts ...InstrumentOption) UpDownCounter

    // Histogram creates a histogram
    Histogram(name string, opts ...InstrumentOption) Histogram

    // Gauge creates a gauge (async)
    Gauge(name string, callback func() float64, opts ...InstrumentOption) error
}

// Counter is a monotonically increasing value
type Counter interface {
    Add(ctx context.Context, incr int64, attrs ...Attribute)
}

// UpDownCounter can increase or decrease
type UpDownCounter interface {
    Add(ctx context.Context, incr int64, attrs ...Attribute)
}

// Histogram records value distributions
type Histogram interface {
    Record(ctx context.Context, value float64, attrs ...Attribute)
}

// Attribute is a key-value pair
type Attribute struct {
    Key   string
    Value interface{}
}
```

## Configuration

```go
// Config holds observability configuration
type Config struct {
    // Service name
    ServiceName string `env:"OTEL_SERVICE_NAME" required:"true"`

    // Service version
    ServiceVersion string `env:"OTEL_SERVICE_VERSION" default:""`

    // Environment (production, staging, development)
    Environment string `env:"OTEL_ENVIRONMENT" default:"development"`

    // Tracing configuration
    Tracing TracingConfig

    // Metrics configuration
    Metrics MetricsConfig
}

type TracingConfig struct {
    // Enable tracing
    Enabled bool `env:"OTEL_TRACING_ENABLED" default:"true"`

    // Exporter: "jaeger", "zipkin", "otlp", "stdout", "none"
    Exporter string `env:"OTEL_TRACES_EXPORTER" default:"otlp"`

    // Sampling rate (0.0 to 1.0)
    SamplingRate float64 `env:"OTEL_TRACES_SAMPLER_ARG" default:"1.0"`

    // Sampler type: "always_on", "always_off", "traceidratio", "parentbased_always_on"
    Sampler string `env:"OTEL_TRACES_SAMPLER" default:"parentbased_always_on"`

    // Batch span processor settings
    BatchTimeout time.Duration `env:"OTEL_BSP_SCHEDULE_DELAY" default:"5s"`
    MaxQueueSize int           `env:"OTEL_BSP_MAX_QUEUE_SIZE" default:"2048"`
    MaxExportBatchSize int     `env:"OTEL_BSP_MAX_EXPORT_BATCH_SIZE" default:"512"`
}

type MetricsConfig struct {
    // Enable metrics
    Enabled bool `env:"OTEL_METRICS_ENABLED" default:"true"`

    // Exporter: "prometheus", "otlp", "stdout", "none"
    Exporter string `env:"OTEL_METRICS_EXPORTER" default:"prometheus"`

    // Prometheus HTTP path
    PrometheusPath string `env:"OTEL_PROMETHEUS_PATH" default:"/metrics"`

    // Export interval for push-based exporters
    ExportInterval time.Duration `env:"OTEL_METRIC_EXPORT_INTERVAL" default:"60s"`
}
```

## Exporter Configurations

### OTLP (OpenTelemetry Protocol)

```go
type OTLPConfig struct {
    // Endpoint for traces
    TracesEndpoint string `env:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT" default:"localhost:4317"`

    // Endpoint for metrics
    MetricsEndpoint string `env:"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT" default:"localhost:4317"`

    // Protocol: "grpc" or "http/protobuf"
    Protocol string `env:"OTEL_EXPORTER_OTLP_PROTOCOL" default:"grpc"`

    // Headers for authentication
    Headers map[string]string `env:"OTEL_EXPORTER_OTLP_HEADERS"`

    // TLS configuration
    Insecure bool `env:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`

    // Compression: "gzip" or "none"
    Compression string `env:"OTEL_EXPORTER_OTLP_COMPRESSION" default:"gzip"`

    // Timeout
    Timeout time.Duration `env:"OTEL_EXPORTER_OTLP_TIMEOUT" default:"10s"`
}
```

### Jaeger

```go
type JaegerConfig struct {
    // Agent endpoint (UDP)
    AgentHost string `env:"OTEL_EXPORTER_JAEGER_AGENT_HOST" default:"localhost"`
    AgentPort int    `env:"OTEL_EXPORTER_JAEGER_AGENT_PORT" default:"6831"`

    // Collector endpoint (HTTP)
    CollectorEndpoint string `env:"OTEL_EXPORTER_JAEGER_ENDPOINT" default:""`

    // Username/password for collector
    Username string `env:"OTEL_EXPORTER_JAEGER_USER" default:""`
    Password string `env:"OTEL_EXPORTER_JAEGER_PASSWORD" default:""`
}
```

### Prometheus

```go
type PrometheusConfig struct {
    // HTTP path for metrics endpoint
    Path string `env:"OTEL_PROMETHEUS_PATH" default:"/metrics"`

    // Enable default Go runtime metrics
    GoMetrics bool `env:"OTEL_PROMETHEUS_GO_METRICS" default:"true"`

    // Enable process metrics
    ProcessMetrics bool `env:"OTEL_PROMETHEUS_PROCESS_METRICS" default:"true"`

    // Namespace prefix for metrics
    Namespace string `env:"OTEL_PROMETHEUS_NAMESPACE" default:""`
}
```

## Span Options

```go
// SpanOption configures span creation
type SpanOption func(*spanConfig)

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption

// WithAttributes sets initial attributes
func WithAttributes(attrs ...Attribute) SpanOption

// WithLinks sets span links
func WithLinks(links ...Link) SpanOption

// WithStartTime sets a custom start time
func WithStartTime(t time.Time) SpanOption

// SpanKind describes the relationship between spans
type SpanKind int

const (
    SpanKindInternal SpanKind = iota
    SpanKindServer
    SpanKindClient
    SpanKindProducer
    SpanKindConsumer
)
```

## Helper Functions

```go
// StartSpan is a convenience function for starting spans
func StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

// SpanFromContext extracts span from context
func SpanFromContext(ctx context.Context) Span

// TraceIDFromContext extracts trace ID from context
func TraceIDFromContext(ctx context.Context) string

// SpanIDFromContext extracts span ID from context
func SpanIDFromContext(ctx context.Context) string

// Attr creates an attribute
func Attr(key string, value interface{}) Attribute

// String creates a string attribute
func String(key, value string) Attribute

// Int creates an int attribute
func Int(key string, value int) Attribute

// Int64 creates an int64 attribute
func Int64(key string, value int64) Attribute

// Float64 creates a float64 attribute
func Float64(key string, value float64) Attribute

// Bool creates a bool attribute
func Bool(key string, value bool) Attribute

// StringSlice creates a string slice attribute
func StringSlice(key string, value []string) Attribute
```

## HTTP Middleware

```go
package middleware

// HTTP returns HTTP server middleware
func HTTP(opts ...Option) func(http.Handler) http.Handler

// Option configures the middleware
type Option func(*config)

// WithTracerProvider sets custom tracer provider
func WithTracerProvider(tp Provider) Option

// WithPropagators sets custom propagators
func WithPropagators(propagators propagation.TextMapPropagator) Option

// WithSpanNameFormatter customizes span names
func WithSpanNameFormatter(fn func(*http.Request) string) Option

// WithFilter skips tracing for certain requests
func WithFilter(fn func(*http.Request) bool) Option

// Default middleware behavior:
// - Creates span for each request
// - Extracts trace context from headers
// - Records HTTP attributes (method, URL, status)
// - Records errors
// - Propagates context to downstream services
```

## gRPC Interceptors

```go
package middleware

// UnaryServerInterceptor returns unary server interceptor
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor

// StreamServerInterceptor returns stream server interceptor
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor

// UnaryClientInterceptor returns unary client interceptor
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor

// StreamClientInterceptor returns stream client interceptor
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor
```

## HTTP Client Instrumentation

```go
package middleware

// HTTPClientTransport wraps http.RoundTripper with tracing
func HTTPClientTransport(base http.RoundTripper, opts ...Option) http.RoundTripper

// Usage:
client := &http.Client{
    Transport: middleware.HTTPClientTransport(http.DefaultTransport),
}
```

## Logging Integration

```go
package logging

// Fields returns trace-correlated log fields
func Fields(ctx context.Context) map[string]interface{} {
    return map[string]interface{}{
        "trace_id": observability.TraceIDFromContext(ctx),
        "span_id":  observability.SpanIDFromContext(ctx),
    }
}

// LoggerAdapter adapts popular loggers to include trace context
type LoggerAdapter interface {
    WithTraceContext(ctx context.Context) Logger
}

// Example with zerolog:
log.Info().
    Str("trace_id", observability.TraceIDFromContext(ctx)).
    Str("span_id", observability.SpanIDFromContext(ctx)).
    Msg("Processing request")
```

## Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/observability"
)

func main() {
    // Initialize observability
    provider, err := observability.New(observability.Config{
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        Environment:    "production",
        Tracing: observability.TracingConfig{
            Enabled:      true,
            Exporter:     "otlp",
            SamplingRate: 0.1, // 10% sampling
        },
        Metrics: observability.MetricsConfig{
            Enabled:  true,
            Exporter: "prometheus",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())

    // Create tracer and meter
    tracer := provider.Tracer("my-service")
    meter := provider.Meter("my-service")

    // Start application
    runApp(tracer, meter)
}
```

### Creating Spans

```go
func processOrder(ctx context.Context, orderID string) error {
    ctx, span := observability.StartSpan(ctx, "processOrder",
        observability.WithSpanKind(observability.SpanKindInternal),
        observability.WithAttributes(
            observability.String("order.id", orderID),
        ),
    )
    defer span.End()

    // Validate order
    if err := validateOrder(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(observability.StatusError, err.Error())
        return err
    }

    // Process payment
    span.AddEvent("processing_payment")
    if err := processPayment(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(observability.StatusError, err.Error())
        return err
    }

    span.SetStatus(observability.StatusOK, "order processed")
    return nil
}
```

### Recording Metrics

```go
func main() {
    provider, _ := observability.New(cfg)
    meter := provider.Meter("my-service")

    // Create instruments
    requestCounter := meter.Counter("http_requests_total",
        observability.WithDescription("Total HTTP requests"),
        observability.WithUnit("1"),
    )

    requestDuration := meter.Histogram("http_request_duration_seconds",
        observability.WithDescription("HTTP request duration"),
        observability.WithUnit("s"),
        observability.WithBuckets([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}),
    )

    activeRequests := meter.UpDownCounter("http_requests_active",
        observability.WithDescription("Active HTTP requests"),
    )

    // Record metrics
    ctx := context.Background()

    requestCounter.Add(ctx, 1,
        observability.String("method", "GET"),
        observability.String("path", "/api/users"),
        observability.Int("status", 200),
    )

    requestDuration.Record(ctx, 0.125,
        observability.String("method", "GET"),
    )
}
```

### HTTP Server with Middleware

```go
import (
    "github.com/user/core-backend/pkg/observability"
    "github.com/user/core-backend/pkg/observability/middleware"
)

func main() {
    provider, _ := observability.New(cfg)

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)

    // Apply middleware
    handler := middleware.HTTP(
        middleware.WithFilter(func(r *http.Request) bool {
            // Skip health checks
            return r.URL.Path != "/health"
        }),
    )(mux)

    http.ListenAndServe(":8080", handler)
}
```

### gRPC Server with Interceptors

```go
import (
    "github.com/user/core-backend/pkg/observability/middleware"
)

func main() {
    provider, _ := observability.New(cfg)

    server := grpc.NewServer(
        grpc.UnaryInterceptor(middleware.UnaryServerInterceptor()),
        grpc.StreamInterceptor(middleware.StreamServerInterceptor()),
    )

    pb.RegisterUserServiceServer(server, &userService{})
    server.Serve(lis)
}
```

### Propagating Context Across Services

```go
func callDownstreamService(ctx context.Context) error {
    // Create instrumented HTTP client
    client := &http.Client{
        Transport: middleware.HTTPClientTransport(http.DefaultTransport),
    }

    req, _ := http.NewRequestWithContext(ctx, "GET", "http://other-service/api", nil)

    // Trace context is automatically propagated via headers
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### With Core-Backend Server Package

```go
import (
    "github.com/user/core-backend/pkg/server"
    "github.com/user/core-backend/pkg/observability"
)

func main() {
    obs, _ := observability.New(cfg)

    srv := server.New(
        server.WithObservability(obs),
        // Automatically adds HTTP and gRPC instrumentation
    )

    srv.Start()
}
```

## Prometheus Metrics Handler

```go
import (
    "github.com/user/core-backend/pkg/observability"
)

func main() {
    provider, _ := observability.New(cfg)

    // Get Prometheus HTTP handler
    metricsHandler := provider.PrometheusHandler()

    mux := http.NewServeMux()
    mux.Handle("/metrics", metricsHandler)

    http.ListenAndServe(":9090", mux)
}
```

## Health Check

```go
// HealthCheck verifies observability backend connectivity
func (p *Provider) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        // Verify exporter connectivity
        return p.ForceFlush(ctx)
    }
}
```

## Dependencies

- **Required:**
  - `go.opentelemetry.io/otel` - Core OpenTelemetry API
  - `go.opentelemetry.io/otel/sdk` - OpenTelemetry SDK
- **Optional:**
  - `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` for OTLP
  - `go.opentelemetry.io/otel/exporters/jaeger` for Jaeger
  - `go.opentelemetry.io/otel/exporters/zipkin` for Zipkin
  - `go.opentelemetry.io/otel/exporters/prometheus` for Prometheus

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with in-memory exporter
- Benchmark tests for hot paths
- Middleware tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Provider
1. Define Provider, Tracer, Meter interfaces
2. Implement provider initialization
3. Configuration loading
4. Basic span creation

### Phase 2: Exporters
1. OTLP exporter (primary)
2. Jaeger exporter
3. Prometheus exporter
4. Stdout exporter for development

### Phase 3: Middleware
1. HTTP server middleware
2. gRPC interceptors
3. HTTP client instrumentation

### Phase 4: Metrics
1. Counter, Histogram, Gauge instruments
2. Default metrics (runtime, process)
3. Prometheus handler

### Phase 5: Logging Integration
1. Trace context extraction
2. Logger adapters
3. Structured logging examples

### Phase 6: Documentation & Examples
1. README with full documentation
2. Basic setup example
3. HTTP server example
4. Microservices tracing example
