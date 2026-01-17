# Package Plan: pkg/queue

## Overview

A unified message queue abstraction supporting multiple backends (Kafka, RabbitMQ, NATS, in-memory). Provides a consistent API for publishing and consuming messages with support for acknowledgment, dead letter queues, and retries.

## Goals

1. **Unified Interface** - Single API for all queue backends
2. **Multiple Backends** - Kafka, RabbitMQ, NATS, and in-memory
3. **Reliability** - At-least-once delivery with acknowledgments
4. **Dead Letter Queues** - Automatic routing of failed messages
5. **Retry Mechanism** - Configurable retry with exponential backoff
6. **Consumer Groups** - Support for distributed consumption
7. **Serialization** - Pluggable message encoding
8. **Zero Core Dependencies** - Backend drivers are optional

## Architecture

```
pkg/queue/
├── queue.go              # Core interfaces (Publisher, Consumer, Message)
├── config.go             # Configuration with env support
├── options.go            # Functional options
├── message.go            # Message type with headers/metadata
├── errors.go             # Custom error types
├── memory.go             # In-memory implementation (for testing/dev)
├── memory_test.go
├── kafka/
│   ├── kafka.go          # Kafka implementation
│   ├── config.go         # Kafka-specific config
│   └── kafka_test.go
├── rabbitmq/
│   ├── rabbitmq.go       # RabbitMQ implementation
│   ├── config.go         # RabbitMQ-specific config
│   └── rabbitmq_test.go
├── nats/
│   ├── nats.go           # NATS implementation
│   ├── config.go         # NATS-specific config
│   └── nats_test.go
├── middleware/
│   ├── logging.go        # Logging middleware
│   ├── retry.go          # Retry middleware
│   ├── deadletter.go     # Dead letter queue middleware
│   └── tracing.go        # Tracing middleware
├── examples/
│   ├── basic/
│   ├── kafka/
│   ├── rabbitmq/
│   └── consumer-group/
└── README.md
```

## Core Interfaces

```go
package queue

import (
    "context"
    "time"
)

// Message represents a queue message
type Message struct {
    // ID is a unique identifier for the message
    ID string

    // Topic/Queue name
    Topic string

    // Body is the message payload
    Body []byte

    // Headers are key-value metadata
    Headers map[string]string

    // Timestamp when message was published
    Timestamp time.Time

    // DeliveryCount tracks retry attempts
    DeliveryCount int

    // Raw holds backend-specific message data
    Raw interface{}
}

// Publisher publishes messages to queues
type Publisher interface {
    // Publish sends a message to the specified topic
    Publish(ctx context.Context, topic string, msg *Message) error

    // PublishBatch sends multiple messages
    PublishBatch(ctx context.Context, topic string, msgs []*Message) error

    // Close releases resources
    Close() error
}

// Consumer consumes messages from queues
type Consumer interface {
    // Subscribe starts consuming from topics
    Subscribe(ctx context.Context, topics []string, handler Handler) error

    // Close stops consuming and releases resources
    Close() error
}

// Handler processes consumed messages
type Handler interface {
    Handle(ctx context.Context, msg *Message) error
}

// HandlerFunc is a function adapter for Handler
type HandlerFunc func(ctx context.Context, msg *Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg *Message) error {
    return f(ctx, msg)
}

// Acknowledger allows explicit message acknowledgment
type Acknowledger interface {
    Ack(ctx context.Context) error
    Nack(ctx context.Context, requeue bool) error
}
```

## Configuration

```go
// Config holds queue configuration
type Config struct {
    // Backend type: "memory", "kafka", "rabbitmq", "nats"
    Backend string `env:"QUEUE_BACKEND" default:"memory"`

    // Consumer group ID for distributed consumption
    ConsumerGroup string `env:"QUEUE_CONSUMER_GROUP" default:""`

    // Concurrency for message processing
    Concurrency int `env:"QUEUE_CONCURRENCY" default:"1"`

    // Retry configuration
    Retry RetryConfig

    // Dead letter queue configuration
    DeadLetter DeadLetterConfig
}

type RetryConfig struct {
    // Maximum retry attempts (0 = no retry)
    MaxAttempts int `env:"QUEUE_RETRY_MAX_ATTEMPTS" default:"3"`

    // Initial delay between retries
    InitialDelay time.Duration `env:"QUEUE_RETRY_INITIAL_DELAY" default:"1s"`

    // Maximum delay between retries
    MaxDelay time.Duration `env:"QUEUE_RETRY_MAX_DELAY" default:"30s"`

    // Multiplier for exponential backoff
    Multiplier float64 `env:"QUEUE_RETRY_MULTIPLIER" default:"2.0"`
}

type DeadLetterConfig struct {
    // Enable dead letter queue
    Enabled bool `env:"QUEUE_DLQ_ENABLED" default:"true"`

    // Topic suffix for dead letter queues
    TopicSuffix string `env:"QUEUE_DLQ_SUFFIX" default:".dlq"`
}
```

## Backend Configurations

### Kafka

```go
type KafkaConfig struct {
    // Broker addresses
    Brokers []string `env:"KAFKA_BROKERS" default:"localhost:9092"`

    // Client ID
    ClientID string `env:"KAFKA_CLIENT_ID" default:""`

    // SASL authentication
    SASL struct {
        Enabled   bool   `env:"KAFKA_SASL_ENABLED" default:"false"`
        Mechanism string `env:"KAFKA_SASL_MECHANISM" default:"PLAIN"`
        Username  string `env:"KAFKA_SASL_USERNAME"`
        Password  string `env:"KAFKA_SASL_PASSWORD"`
    }

    // TLS configuration
    TLS struct {
        Enabled            bool   `env:"KAFKA_TLS_ENABLED" default:"false"`
        CertFile           string `env:"KAFKA_TLS_CERT_FILE"`
        KeyFile            string `env:"KAFKA_TLS_KEY_FILE"`
        CAFile             string `env:"KAFKA_TLS_CA_FILE"`
        InsecureSkipVerify bool   `env:"KAFKA_TLS_INSECURE" default:"false"`
    }

    // Producer settings
    Producer struct {
        Acks           string        `env:"KAFKA_PRODUCER_ACKS" default:"all"`
        Timeout        time.Duration `env:"KAFKA_PRODUCER_TIMEOUT" default:"10s"`
        BatchSize      int           `env:"KAFKA_PRODUCER_BATCH_SIZE" default:"16384"`
        Compression    string        `env:"KAFKA_PRODUCER_COMPRESSION" default:"snappy"`
        Idempotent     bool          `env:"KAFKA_PRODUCER_IDEMPOTENT" default:"true"`
    }

    // Consumer settings
    Consumer struct {
        AutoOffset     string        `env:"KAFKA_CONSUMER_AUTO_OFFSET" default:"earliest"`
        SessionTimeout time.Duration `env:"KAFKA_CONSUMER_SESSION_TIMEOUT" default:"10s"`
        HeartbeatInterval time.Duration `env:"KAFKA_CONSUMER_HEARTBEAT_INTERVAL" default:"3s"`
    }
}
```

### RabbitMQ

```go
type RabbitMQConfig struct {
    // Connection URL
    URL string `env:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`

    // Exchange settings
    Exchange struct {
        Name    string `env:"RABBITMQ_EXCHANGE_NAME" default:""`
        Type    string `env:"RABBITMQ_EXCHANGE_TYPE" default:"topic"`
        Durable bool   `env:"RABBITMQ_EXCHANGE_DURABLE" default:"true"`
    }

    // Queue settings
    Queue struct {
        Durable    bool          `env:"RABBITMQ_QUEUE_DURABLE" default:"true"`
        AutoDelete bool          `env:"RABBITMQ_QUEUE_AUTO_DELETE" default:"false"`
        TTL        time.Duration `env:"RABBITMQ_QUEUE_TTL" default:"0"`
    }

    // Prefetch count for consumer
    PrefetchCount int `env:"RABBITMQ_PREFETCH_COUNT" default:"10"`

    // Connection recovery
    ReconnectDelay time.Duration `env:"RABBITMQ_RECONNECT_DELAY" default:"5s"`
}
```

### NATS

```go
type NATSConfig struct {
    // Server URLs
    Servers []string `env:"NATS_SERVERS" default:"nats://localhost:4222"`

    // Cluster name (for NATS Streaming/JetStream)
    ClusterID string `env:"NATS_CLUSTER_ID" default:""`

    // Credentials
    Username string `env:"NATS_USERNAME" default:""`
    Password string `env:"NATS_PASSWORD" default:""`
    Token    string `env:"NATS_TOKEN" default:""`

    // TLS
    TLS struct {
        Enabled  bool   `env:"NATS_TLS_ENABLED" default:"false"`
        CertFile string `env:"NATS_TLS_CERT_FILE"`
        KeyFile  string `env:"NATS_TLS_KEY_FILE"`
        CAFile   string `env:"NATS_TLS_CA_FILE"`
    }

    // JetStream configuration
    JetStream struct {
        Enabled     bool   `env:"NATS_JETSTREAM_ENABLED" default:"false"`
        StorageType string `env:"NATS_JETSTREAM_STORAGE" default:"file"`
        Replicas    int    `env:"NATS_JETSTREAM_REPLICAS" default:"1"`
    }
}
```

## In-Memory Implementation

```go
// Memory implements Publisher and Consumer for testing/development
type Memory struct {
    topics     map[string]chan *Message
    handlers   map[string][]Handler
    mu         sync.RWMutex
    bufferSize int
}

// NewMemory creates an in-memory queue
func NewMemory(opts ...Option) *Memory

// Options
func WithBufferSize(size int) Option
```

## Middleware

```go
// Middleware wraps a Handler
type Middleware func(Handler) Handler

// Chain combines multiple middleware
func Chain(middlewares ...Middleware) Middleware

// Logging middleware
func Logging(logger Logger) Middleware

// Retry middleware with exponential backoff
func Retry(cfg RetryConfig) Middleware

// DeadLetter routes failed messages
func DeadLetter(publisher Publisher, cfg DeadLetterConfig) Middleware

// Tracing adds distributed tracing
func Tracing(tracer Tracer) Middleware

// Recovery recovers from panics
func Recovery(logger Logger) Middleware
```

## Error Handling

```go
var (
    // ErrClosed is returned when operating on closed queue
    ErrClosed = errors.New("queue: connection closed")

    // ErrTimeout is returned on operation timeout
    ErrTimeout = errors.New("queue: operation timeout")

    // ErrTopicNotFound is returned when topic doesn't exist
    ErrTopicNotFound = errors.New("queue: topic not found")

    // ErrMessageRejected is returned when message is rejected
    ErrMessageRejected = errors.New("queue: message rejected")

    // ErrMaxRetriesExceeded is returned after max retry attempts
    ErrMaxRetriesExceeded = errors.New("queue: max retries exceeded")
)

// Retryable checks if error is retryable
func Retryable(err error) bool
```

## Usage Examples

### Basic Publishing

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/queue"
)

func main() {
    // Create publisher
    pub, err := queue.NewPublisher(queue.Config{
        Backend: "memory",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer pub.Close()

    ctx := context.Background()

    // Publish a message
    msg := &queue.Message{
        Body: []byte(`{"user_id": "123", "action": "signup"}`),
        Headers: map[string]string{
            "content-type": "application/json",
        },
    }

    err = pub.Publish(ctx, "user.events", msg)
}
```

### Basic Consuming

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/queue"
)

func main() {
    // Create consumer
    cons, err := queue.NewConsumer(queue.Config{
        Backend:       "memory",
        ConsumerGroup: "my-service",
        Concurrency:   5,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer cons.Close()

    ctx := context.Background()

    // Subscribe with handler
    err = cons.Subscribe(ctx, []string{"user.events"}, queue.HandlerFunc(
        func(ctx context.Context, msg *queue.Message) error {
            log.Printf("Received: %s", msg.Body)
            return nil
        },
    ))
}
```

### Kafka Producer/Consumer

```go
import (
    "github.com/user/core-backend/pkg/queue"
    "github.com/user/core-backend/pkg/queue/kafka"
)

func main() {
    cfg := kafka.Config{
        Brokers:  []string{"kafka1:9092", "kafka2:9092"},
        ClientID: "my-service",
    }

    // Create Kafka publisher
    pub, err := kafka.NewPublisher(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Create Kafka consumer with group
    cons, err := kafka.NewConsumer(cfg, queue.Config{
        ConsumerGroup: "my-service-group",
        Concurrency:   10,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Use same interface as memory queue
    pub.Publish(ctx, "events", msg)
    cons.Subscribe(ctx, []string{"events"}, handler)
}
```

### With Middleware

```go
func main() {
    cons, _ := queue.NewConsumer(cfg)

    // Create handler with middleware chain
    handler := queue.Chain(
        queue.Recovery(logger),
        queue.Logging(logger),
        queue.Retry(queue.RetryConfig{
            MaxAttempts:  3,
            InitialDelay: time.Second,
        }),
        queue.DeadLetter(publisher, queue.DeadLetterConfig{
            Enabled:     true,
            TopicSuffix: ".dlq",
        }),
    )(myHandler)

    cons.Subscribe(ctx, topics, handler)
}
```

### Batch Publishing

```go
func main() {
    pub, _ := queue.NewPublisher(cfg)

    messages := []*queue.Message{
        {Body: []byte("msg1")},
        {Body: []byte("msg2")},
        {Body: []byte("msg3")},
    }

    // Publish batch atomically
    err := pub.PublishBatch(ctx, "events", messages)
}
```

## Typed Messages

```go
// TypedPublisher provides type-safe publishing
type TypedPublisher[T any] struct {
    publisher  Publisher
    topic      string
    serializer Serializer
}

func NewTypedPublisher[T any](pub Publisher, topic string, ser Serializer) *TypedPublisher[T]

func (p *TypedPublisher[T]) Publish(ctx context.Context, payload T) error

// Usage
type UserEvent struct {
    UserID string `json:"user_id"`
    Action string `json:"action"`
}

pub := queue.NewTypedPublisher[UserEvent](publisher, "user.events", queue.JSONSerializer{})
pub.Publish(ctx, UserEvent{UserID: "123", Action: "signup"})
```

## Health Check

```go
// HealthCheck returns a health check function
func (c *Consumer) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        // Check connection to broker
        return c.ping(ctx)
    }
}
```

## Observability Hooks

```go
// Hook interface for observability
type Hook interface {
    BeforePublish(ctx context.Context, topic string, msg *Message)
    AfterPublish(ctx context.Context, topic string, msg *Message, err error)
    BeforeHandle(ctx context.Context, msg *Message)
    AfterHandle(ctx context.Context, msg *Message, err error)
}

// WithHook adds observability hooks
func WithHook(hook Hook) Option
```

## Dependencies

- **Required:** None (in-memory implementation)
- **Optional:**
  - `github.com/segmentio/kafka-go` or `github.com/confluentinc/confluent-kafka-go` for Kafka
  - `github.com/rabbitmq/amqp091-go` for RabbitMQ
  - `github.com/nats-io/nats.go` for NATS

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with testcontainers
- Benchmark tests for throughput
- Race condition tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Memory Implementation
1. Define Publisher, Consumer, Message interfaces
2. Implement in-memory queue
3. Add middleware system
4. Write comprehensive tests

### Phase 2: Kafka Implementation
1. Implement Kafka producer/consumer
2. Add partition support
3. Consumer group handling
4. Integration tests

### Phase 3: RabbitMQ Implementation
1. Implement RabbitMQ publisher/consumer
2. Exchange/queue bindings
3. Message acknowledgment
4. Connection recovery

### Phase 4: NATS Implementation
1. Implement NATS core
2. JetStream support
3. Durable subscriptions

### Phase 5: Advanced Features
1. Dead letter queue middleware
2. Retry middleware
3. Tracing middleware
4. Typed message helpers

### Phase 6: Documentation & Examples
1. README with full documentation
2. Example for each backend
3. Consumer group example
4. Middleware chain example
