# Package Plan: pkg/pubsub

## Overview

A publish-subscribe abstraction for event broadcasting between services. Supports multiple backends (Redis, Google Pub/Sub, AWS SNS/SQS) with features like message filtering, fan-out, and exactly-once delivery.

## Goals

1. **Multiple Backends** - Redis, Google Pub/Sub, AWS SNS/SQS, NATS
2. **Topic-Based** - Publish to topics, subscribe by pattern
3. **Message Filtering** - Attribute-based filtering
4. **Fan-Out** - Multiple subscribers per topic
5. **Acknowledgment** - At-least-once delivery
6. **Dead Letter** - Handle failed messages
7. **Ordering** - Optional message ordering

## Architecture

```
pkg/pubsub/
├── pubsub.go             # Core interfaces
├── config.go             # Configuration
├── options.go            # Functional options
├── message.go            # Message definition
├── topic.go              # Topic management
├── subscription.go       # Subscription management
├── provider/
│   ├── provider.go       # Provider interface
│   ├── memory.go         # In-memory (testing)
│   ├── redis.go          # Redis Pub/Sub
│   ├── gcp.go            # Google Pub/Sub
│   ├── aws.go            # AWS SNS/SQS
│   └── nats.go           # NATS
├── middleware/
│   ├── logging.go        # Logging
│   ├── tracing.go        # Distributed tracing
│   └── retry.go          # Retry logic
├── examples/
│   ├── basic/
│   ├── redis/
│   └── gcp/
└── README.md
```

## Core Interfaces

```go
package pubsub

import (
    "context"
    "time"
)

// Client provides pub/sub functionality
type Client interface {
    // Topic returns a topic by name
    Topic(name string) Topic

    // CreateTopic creates a new topic
    CreateTopic(ctx context.Context, name string, opts ...TopicOption) (Topic, error)

    // Subscription returns a subscription by name
    Subscription(name string) Subscription

    // CreateSubscription creates a new subscription
    CreateSubscription(ctx context.Context, name string, topic Topic, opts ...SubscriptionOption) (Subscription, error)

    // Close releases resources
    Close() error
}

// Topic represents a pub/sub topic
type Topic interface {
    // Publish publishes a message
    Publish(ctx context.Context, msg *Message) (string, error)

    // PublishAsync publishes asynchronously
    PublishAsync(ctx context.Context, msg *Message) PublishResult

    // Exists checks if topic exists
    Exists(ctx context.Context) (bool, error)

    // Delete deletes the topic
    Delete(ctx context.Context) error

    // ID returns the topic ID
    ID() string
}

// Subscription represents a topic subscription
type Subscription interface {
    // Receive starts receiving messages
    Receive(ctx context.Context, handler MessageHandler) error

    // Exists checks if subscription exists
    Exists(ctx context.Context) (bool, error)

    // Delete deletes the subscription
    Delete(ctx context.Context) error

    // ID returns the subscription ID
    ID() string
}

// Message represents a pub/sub message
type Message struct {
    // ID is the message identifier
    ID string

    // Data is the message payload
    Data []byte

    // Attributes are key-value metadata
    Attributes map[string]string

    // PublishTime is when the message was published
    PublishTime time.Time

    // OrderingKey for ordered delivery
    OrderingKey string

    // DeliveryAttempt count
    DeliveryAttempt int

    // ack/nack functions (set by subscription)
    ack  func()
    nack func()
}

// Ack acknowledges the message
func (m *Message) Ack()

// Nack negatively acknowledges (requeue)
func (m *Message) Nack()

// MessageHandler processes messages
type MessageHandler func(ctx context.Context, msg *Message) error

// PublishResult for async publish
type PublishResult interface {
    Get(ctx context.Context) (string, error)
}
```

## Configuration

```go
// Config holds pub/sub configuration
type Config struct {
    // Provider: "memory", "redis", "gcp", "aws", "nats"
    Provider string `env:"PUBSUB_PROVIDER" default:"memory"`

    // Project ID (for GCP)
    ProjectID string `env:"PUBSUB_PROJECT_ID"`
}

// TopicOption configures a topic
type TopicOption func(*topicConfig)

// WithMessageOrdering enables message ordering
func WithMessageOrdering() TopicOption

// WithRetentionDuration sets message retention
func WithRetentionDuration(d time.Duration) TopicOption

// SubscriptionOption configures a subscription
type SubscriptionOption func(*subscriptionConfig)

// WithAckDeadline sets acknowledgment deadline
func WithAckDeadline(d time.Duration) SubscriptionOption

// WithFilter sets message filter
func WithFilter(filter string) SubscriptionOption

// WithDeadLetter configures dead letter topic
func WithDeadLetter(topic Topic, maxAttempts int) SubscriptionOption

// WithMaxConcurrency sets max concurrent handlers
func WithMaxConcurrency(n int) SubscriptionOption
```

## Provider Configurations

```go
// RedisConfig for Redis Pub/Sub
type RedisConfig struct {
    URL       string `env:"PUBSUB_REDIS_URL" default:"redis://localhost:6379"`
    KeyPrefix string `env:"PUBSUB_REDIS_PREFIX" default:"pubsub:"`
}

// GCPConfig for Google Pub/Sub
type GCPConfig struct {
    ProjectID       string `env:"GCP_PROJECT_ID" required:"true"`
    CredentialsFile string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
}

// AWSConfig for AWS SNS/SQS
type AWSConfig struct {
    Region          string `env:"AWS_REGION" default:"us-east-1"`
    AccountID       string `env:"AWS_ACCOUNT_ID"`
    AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
    SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
}
```

## Usage Examples

### Basic Publishing

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/pubsub"
)

func main() {
    client, err := pubsub.New(pubsub.Config{
        Provider: "redis",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Get or create topic
    topic, _ := client.CreateTopic(ctx, "user-events")

    // Publish message
    msg := &pubsub.Message{
        Data: []byte(`{"event": "user.created", "user_id": "123"}`),
        Attributes: map[string]string{
            "event_type": "user.created",
        },
    }

    id, err := topic.Publish(ctx, msg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Published message: %s", id)
}
```

### Subscribing

```go
func main() {
    client, _ := pubsub.New(cfg)
    defer client.Close()

    ctx := context.Background()

    topic, _ := client.CreateTopic(ctx, "user-events")

    // Create subscription
    sub, err := client.CreateSubscription(ctx, "user-service-sub", topic,
        pubsub.WithAckDeadline(30*time.Second),
        pubsub.WithMaxConcurrency(10),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Receive messages
    err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) error {
        log.Printf("Received: %s", msg.Data)

        // Process message
        if err := processMessage(msg); err != nil {
            msg.Nack() // Requeue
            return err
        }

        msg.Ack() // Acknowledge
        return nil
    })
}
```

### With Filtering

```go
func main() {
    client, _ := pubsub.New(cfg)

    topic, _ := client.CreateTopic(ctx, "orders")

    // Subscribe only to high-value orders
    sub, _ := client.CreateSubscription(ctx, "high-value-orders", topic,
        pubsub.WithFilter(`attributes.order_value > 1000`),
    )

    sub.Receive(ctx, handleHighValueOrder)
}

// Publisher sets attributes
topic.Publish(ctx, &pubsub.Message{
    Data: orderJSON,
    Attributes: map[string]string{
        "order_value": "1500",
        "region":      "US",
    },
})
```

### Fan-Out Pattern

```go
func main() {
    client, _ := pubsub.New(cfg)

    // One topic
    topic, _ := client.CreateTopic(ctx, "user-events")

    // Multiple subscribers
    emailSub, _ := client.CreateSubscription(ctx, "email-service", topic)
    analyticsSub, _ := client.CreateSubscription(ctx, "analytics-service", topic)
    auditSub, _ := client.CreateSubscription(ctx, "audit-service", topic)

    // Each subscriber receives all messages
    go emailSub.Receive(ctx, sendEmailNotification)
    go analyticsSub.Receive(ctx, recordAnalytics)
    go auditSub.Receive(ctx, writeAuditLog)
}
```

### Dead Letter Queue

```go
func main() {
    client, _ := pubsub.New(cfg)

    mainTopic, _ := client.CreateTopic(ctx, "orders")
    dlqTopic, _ := client.CreateTopic(ctx, "orders-dlq")

    // Subscription with DLQ
    sub, _ := client.CreateSubscription(ctx, "order-processor", mainTopic,
        pubsub.WithDeadLetter(dlqTopic, 5), // Max 5 attempts
    )

    // Process DLQ separately
    dlqSub, _ := client.CreateSubscription(ctx, "dlq-handler", dlqTopic)
    go dlqSub.Receive(ctx, handleFailedMessage)
}
```

### Ordered Messages

```go
func main() {
    client, _ := pubsub.New(cfg)

    topic, _ := client.CreateTopic(ctx, "user-events",
        pubsub.WithMessageOrdering(),
    )

    // Messages with same ordering key are delivered in order
    topic.Publish(ctx, &pubsub.Message{
        Data:        []byte("event1"),
        OrderingKey: "user-123",
    })

    topic.Publish(ctx, &pubsub.Message{
        Data:        []byte("event2"),
        OrderingKey: "user-123", // Delivered after event1
    })
}
```

### Async Publishing

```go
func main() {
    client, _ := pubsub.New(cfg)
    topic := client.Topic("events")

    var results []pubsub.PublishResult

    // Publish batch asynchronously
    for _, event := range events {
        result := topic.PublishAsync(ctx, &pubsub.Message{
            Data: event,
        })
        results = append(results, result)
    }

    // Wait for all
    for i, result := range results {
        id, err := result.Get(ctx)
        if err != nil {
            log.Printf("Message %d failed: %v", i, err)
        } else {
            log.Printf("Message %d published: %s", i, id)
        }
    }
}
```

## Dependencies

- **Required:** None (memory provider)
- **Optional:**
  - `github.com/redis/go-redis/v9` for Redis
  - `cloud.google.com/go/pubsub` for GCP
  - `github.com/aws/aws-sdk-go-v2` for AWS

## Implementation Phases

### Phase 1: Core Interface & Memory Provider
1. Define Client, Topic, Subscription interfaces
2. In-memory provider
3. Basic pub/sub

### Phase 2: Redis Provider
1. Redis Pub/Sub implementation
2. Pattern subscriptions

### Phase 3: Cloud Providers
1. Google Pub/Sub
2. AWS SNS/SQS

### Phase 4: Advanced Features
1. Message filtering
2. Dead letter queues
3. Message ordering

### Phase 5: Middleware
1. Logging
2. Tracing
3. Retry

### Phase 6: Documentation
1. README
2. Examples
