# Kafka Consumer & Producer Design

## Overview

This document outlines the design for a configurable Kafka event consumer and producer system for the `core-backend` project. The design includes:

- Configurable consumer and producer components
- Dead Letter Queue (DLQ) approach for failed message handling
- Flexible partitioning strategies
- Configurable commit/offset management strategies

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Configuration Design](#configuration-design)
3. [Producer Design](#producer-design)
4. [Consumer Design](#consumer-design)
5. [Dead Letter Queue (DLQ) Design](#dead-letter-queue-dlq-design)
6. [Partitioning Strategies](#partitioning-strategies)
7. [Commit Strategies](#commit-strategies)
8. [Error Handling](#error-handling)
9. [Interfaces & Models](#interfaces--models)
10. [Package Structure](#package-structure)
11. [Usage Examples](#usage-examples)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Application Layer                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────┐              ┌─────────────────┐                       │
│  │   EventHandler  │              │  EventPublisher │                       │
│  │   (Interface)   │              │   (Interface)   │                       │
│  └────────┬────────┘              └────────┬────────┘                       │
│           │                                │                                 │
├───────────┼────────────────────────────────┼─────────────────────────────────┤
│           │         Kafka Package          │                                 │
│           ▼                                ▼                                 │
│  ┌─────────────────┐              ┌─────────────────┐                       │
│  │    Consumer     │              │    Producer     │                       │
│  │  ┌───────────┐  │              │  ┌───────────┐  │                       │
│  │  │Partitioner│  │              │  │Partitioner│  │                       │
│  │  └───────────┘  │              │  └───────────┘  │                       │
│  │  ┌───────────┐  │              │  ┌───────────┐  │                       │
│  │  │ Committer │  │              │  │Serializer │  │                       │
│  │  └───────────┘  │              │  └───────────┘  │                       │
│  │  ┌───────────┐  │              └─────────────────┘                       │
│  │  │DLQHandler │  │                                                        │
│  │  └───────────┘  │                                                        │
│  └─────────────────┘                                                        │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                              Kafka Cluster                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Main Topic   │  │ Main Topic   │  │  DLQ Topic   │  │ Retry Topic  │     │
│  │ Partition 0  │  │ Partition N  │  │              │  │  (Optional)  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Configuration Design

### Environment Variables

Following the existing configuration pattern in the codebase:

```bash
# Kafka Connection
KAFKA_BROKERS=localhost:9092,localhost:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL          # PLAINTEXT, SSL, SASL_PLAINTEXT, SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN                 # PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
KAFKA_SASL_USERNAME=
KAFKA_SASL_PASSWORD=
KAFKA_TLS_CERT_FILE=
KAFKA_TLS_KEY_FILE=
KAFKA_TLS_CA_FILE=
KAFKA_TLS_SKIP_VERIFY=false

# Consumer Configuration
KAFKA_CONSUMER_GROUP_ID=my-service-group
KAFKA_CONSUMER_TOPICS=events,notifications
KAFKA_CONSUMER_AUTO_OFFSET_RESET=earliest  # earliest, latest
KAFKA_CONSUMER_MAX_POLL_RECORDS=500
KAFKA_CONSUMER_MAX_POLL_INTERVAL=5m
KAFKA_CONSUMER_SESSION_TIMEOUT=30s
KAFKA_CONSUMER_HEARTBEAT_INTERVAL=3s
KAFKA_CONSUMER_FETCH_MIN_BYTES=1
KAFKA_CONSUMER_FETCH_MAX_BYTES=52428800
KAFKA_CONSUMER_CONCURRENCY=1               # Number of concurrent consumers

# Commit Strategy
KAFKA_CONSUMER_COMMIT_STRATEGY=auto_commit # auto_commit, sync, async, batch, manual
KAFKA_CONSUMER_AUTO_COMMIT_INTERVAL=5s
KAFKA_CONSUMER_BATCH_COMMIT_SIZE=100
KAFKA_CONSUMER_BATCH_COMMIT_INTERVAL=10s

# Producer Configuration
KAFKA_PRODUCER_ACKS=all                    # 0, 1, all
KAFKA_PRODUCER_RETRIES=3
KAFKA_PRODUCER_RETRY_BACKOFF=100ms
KAFKA_PRODUCER_BATCH_SIZE=16384
KAFKA_PRODUCER_LINGER_MS=5
KAFKA_PRODUCER_COMPRESSION=snappy          # none, gzip, snappy, lz4, zstd
KAFKA_PRODUCER_MAX_REQUEST_SIZE=1048576
KAFKA_PRODUCER_IDEMPOTENT=true
KAFKA_PRODUCER_TIMEOUT=30s

# Partitioning
KAFKA_PRODUCER_PARTITIONER=hash            # hash, round_robin, random, sticky, custom
KAFKA_PRODUCER_PARTITION_KEY_FIELD=id      # Field to use for partition key extraction

# DLQ Configuration
KAFKA_DLQ_ENABLED=true
KAFKA_DLQ_TOPIC_SUFFIX=.dlq
KAFKA_DLQ_MAX_RETRIES=3
KAFKA_DLQ_RETRY_BACKOFF=1s
KAFKA_DLQ_RETRY_BACKOFF_MULTIPLIER=2.0
KAFKA_DLQ_RETRY_MAX_BACKOFF=1m
KAFKA_DLQ_INCLUDE_HEADERS=true
KAFKA_DLQ_RETRY_TOPIC_ENABLED=true
KAFKA_DLQ_RETRY_TOPIC_SUFFIX=.retry
```

### Configuration Struct

```go
// pkg/kafka/config.go

package kafka

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

// Config holds all Kafka configuration.
type Config struct {
    Connection ConnectionConfig `json:"connection"`
    Consumer   ConsumerConfig   `json:"consumer"`
    Producer   ProducerConfig   `json:"producer"`
    DLQ        DLQConfig        `json:"dlq"`
}

// ConnectionConfig holds Kafka broker connection settings.
type ConnectionConfig struct {
    Brokers          []string `json:"brokers"`
    SecurityProtocol string   `json:"security_protocol"`
    SASLMechanism    string   `json:"sasl_mechanism"`
    SASLUsername     string   `json:"sasl_username"`
    SASLPassword     string   `json:"-"` // Don't serialize password
    TLSCertFile      string   `json:"tls_cert_file"`
    TLSKeyFile       string   `json:"tls_key_file"`
    TLSCAFile        string   `json:"tls_ca_file"`
    TLSSkipVerify    bool     `json:"tls_skip_verify"`
}

// ConsumerConfig holds consumer-specific settings.
type ConsumerConfig struct {
    GroupID           string         `json:"group_id"`
    Topics            []string       `json:"topics"`
    AutoOffsetReset   string         `json:"auto_offset_reset"`
    MaxPollRecords    int            `json:"max_poll_records"`
    MaxPollInterval   time.Duration  `json:"max_poll_interval"`
    SessionTimeout    time.Duration  `json:"session_timeout"`
    HeartbeatInterval time.Duration  `json:"heartbeat_interval"`
    FetchMinBytes     int            `json:"fetch_min_bytes"`
    FetchMaxBytes     int            `json:"fetch_max_bytes"`
    Concurrency       int            `json:"concurrency"`
    CommitStrategy    CommitStrategy `json:"commit_strategy"`
    CommitConfig      CommitConfig   `json:"commit_config"`
}

// CommitStrategy defines how offsets are committed.
type CommitStrategy string

const (
    CommitStrategyAutoCommit CommitStrategy = "auto_commit"
    CommitStrategySync       CommitStrategy = "sync"
    CommitStrategyAsync      CommitStrategy = "async"
    CommitStrategyBatch      CommitStrategy = "batch"
    CommitStrategyManual     CommitStrategy = "manual"
)

// CommitConfig holds commit-related settings.
type CommitConfig struct {
    AutoCommitInterval  time.Duration `json:"auto_commit_interval"`
    BatchCommitSize     int           `json:"batch_commit_size"`
    BatchCommitInterval time.Duration `json:"batch_commit_interval"`
}

// ProducerConfig holds producer-specific settings.
type ProducerConfig struct {
    Acks              string             `json:"acks"`
    Retries           int                `json:"retries"`
    RetryBackoff      time.Duration      `json:"retry_backoff"`
    BatchSize         int                `json:"batch_size"`
    LingerMs          int                `json:"linger_ms"`
    Compression       string             `json:"compression"`
    MaxRequestSize    int                `json:"max_request_size"`
    Idempotent        bool               `json:"idempotent"`
    Timeout           time.Duration      `json:"timeout"`
    Partitioner       PartitionerType    `json:"partitioner"`
    PartitionKeyField string             `json:"partition_key_field"`
}

// PartitionerType defines the partitioning strategy.
type PartitionerType string

const (
    PartitionerHash       PartitionerType = "hash"
    PartitionerRoundRobin PartitionerType = "round_robin"
    PartitionerRandom     PartitionerType = "random"
    PartitionerSticky     PartitionerType = "sticky"
    PartitionerCustom     PartitionerType = "custom"
)

// DLQConfig holds Dead Letter Queue settings.
type DLQConfig struct {
    Enabled                bool          `json:"enabled"`
    TopicSuffix            string        `json:"topic_suffix"`
    MaxRetries             int           `json:"max_retries"`
    RetryBackoff           time.Duration `json:"retry_backoff"`
    RetryBackoffMultiplier float64       `json:"retry_backoff_multiplier"`
    RetryMaxBackoff        time.Duration `json:"retry_max_backoff"`
    IncludeHeaders         bool          `json:"include_headers"`
    RetryTopicEnabled      bool          `json:"retry_topic_enabled"`
    RetryTopicSuffix       string        `json:"retry_topic_suffix"`
}
```

---

## Producer Design

### Producer Interface

```go
// pkg/kafka/producer.go

package kafka

import (
    "context"
)

// Producer defines the interface for publishing messages to Kafka.
type Producer interface {
    // Publish sends a single message to the specified topic.
    Publish(ctx context.Context, topic string, message *Message) error

    // PublishBatch sends multiple messages to the specified topic.
    PublishBatch(ctx context.Context, topic string, messages []*Message) error

    // PublishWithKey sends a message with explicit partition key.
    PublishWithKey(ctx context.Context, topic string, key []byte, message *Message) error

    // PublishToPartition sends a message to a specific partition.
    PublishToPartition(ctx context.Context, topic string, partition int32, message *Message) error

    // Close gracefully shuts down the producer.
    Close() error

    // Flush waits for all pending messages to be delivered.
    Flush(ctx context.Context) error
}

// Message represents a Kafka message.
type Message struct {
    Key       []byte            `json:"key,omitempty"`
    Value     []byte            `json:"value"`
    Headers   map[string]string `json:"headers,omitempty"`
    Timestamp time.Time         `json:"timestamp,omitempty"`

    // Metadata for tracking
    Topic     string `json:"topic,omitempty"`
    Partition int32  `json:"partition,omitempty"`
    Offset    int64  `json:"offset,omitempty"`
}

// ProducerOption configures the producer.
type ProducerOption func(*producerOptions)

type producerOptions struct {
    partitioner    Partitioner
    serializer     Serializer
    interceptors   []ProducerInterceptor
    errorHandler   ErrorHandler
    metricsEnabled bool
}

// WithPartitioner sets a custom partitioner.
func WithPartitioner(p Partitioner) ProducerOption {
    return func(o *producerOptions) {
        o.partitioner = p
    }
}

// WithSerializer sets a custom serializer.
func WithSerializer(s Serializer) ProducerOption {
    return func(o *producerOptions) {
        o.serializer = s
    }
}

// WithProducerInterceptor adds an interceptor to the producer.
func WithProducerInterceptor(i ProducerInterceptor) ProducerOption {
    return func(o *producerOptions) {
        o.interceptors = append(o.interceptors, i)
    }
}
```

### Producer Implementation

```go
// pkg/kafka/producer_impl.go

package kafka

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type producer struct {
    config       *ProducerConfig
    client       KafkaClient // Abstracted Kafka client (e.g., confluent-kafka-go, sarama)
    partitioner  Partitioner
    serializer   Serializer
    interceptors []ProducerInterceptor
    errorHandler ErrorHandler
    metrics      *ProducerMetrics

    mu     sync.RWMutex
    closed bool
}

// NewProducer creates a new Kafka producer with the given configuration.
func NewProducer(cfg *Config, opts ...ProducerOption) (Producer, error) {
    options := &producerOptions{
        partitioner: newPartitioner(cfg.Producer.Partitioner),
        serializer:  &JSONSerializer{},
    }

    for _, opt := range opts {
        opt(options)
    }

    client, err := newKafkaProducerClient(cfg)
    if err != nil {
        return nil, fmt.Errorf("creating kafka producer client: %w", err)
    }

    return &producer{
        config:       &cfg.Producer,
        client:       client,
        partitioner:  options.partitioner,
        serializer:   options.serializer,
        interceptors: options.interceptors,
        errorHandler: options.errorHandler,
        metrics:      newProducerMetrics(),
    }, nil
}

func (p *producer) Publish(ctx context.Context, topic string, message *Message) error {
    p.mu.RLock()
    if p.closed {
        p.mu.RUnlock()
        return ErrProducerClosed
    }
    p.mu.RUnlock()

    // Run interceptors
    for _, interceptor := range p.interceptors {
        var err error
        message, err = interceptor.OnSend(ctx, topic, message)
        if err != nil {
            return fmt.Errorf("interceptor error: %w", err)
        }
    }

    // Determine partition
    partition, err := p.partitioner.Partition(topic, message.Key, message.Value)
    if err != nil {
        return fmt.Errorf("partitioning: %w", err)
    }

    // Send message
    if err := p.client.Produce(ctx, topic, partition, message); err != nil {
        p.metrics.RecordError(topic)
        if p.errorHandler != nil {
            p.errorHandler.OnError(ctx, err, message)
        }
        return fmt.Errorf("producing message: %w", err)
    }

    p.metrics.RecordSuccess(topic)
    return nil
}

func (p *producer) PublishBatch(ctx context.Context, topic string, messages []*Message) error {
    var errs []error
    for _, msg := range messages {
        if err := p.Publish(ctx, topic, msg); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return &BatchError{Errors: errs}
    }
    return nil
}

func (p *producer) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.closed {
        return nil
    }
    p.closed = true

    return p.client.Close()
}

func (p *producer) Flush(ctx context.Context) error {
    return p.client.Flush(ctx)
}
```

---

## Consumer Design

### Consumer Interface

```go
// pkg/kafka/consumer.go

package kafka

import (
    "context"
)

// Consumer defines the interface for consuming messages from Kafka.
type Consumer interface {
    // Subscribe registers topics for consumption.
    Subscribe(topics ...string) error

    // Start begins consuming messages, calling the handler for each message.
    Start(ctx context.Context, handler MessageHandler) error

    // Pause temporarily stops consuming from specified partitions.
    Pause(partitions []TopicPartition) error

    // Resume continues consuming from paused partitions.
    Resume(partitions []TopicPartition) error

    // Commit manually commits offsets (when using manual commit strategy).
    Commit(ctx context.Context) error

    // CommitMessage commits offset for a specific message.
    CommitMessage(ctx context.Context, msg *Message) error

    // Close gracefully shuts down the consumer.
    Close() error

    // Lag returns the current consumer lag per partition.
    Lag() map[TopicPartition]int64
}

// TopicPartition represents a topic-partition pair.
type TopicPartition struct {
    Topic     string `json:"topic"`
    Partition int32  `json:"partition"`
}

// MessageHandler processes consumed messages.
type MessageHandler interface {
    // Handle processes a message. Return error to trigger retry/DLQ logic.
    Handle(ctx context.Context, msg *Message) error
}

// MessageHandlerFunc is a function adapter for MessageHandler.
type MessageHandlerFunc func(ctx context.Context, msg *Message) error

func (f MessageHandlerFunc) Handle(ctx context.Context, msg *Message) error {
    return f(ctx, msg)
}

// ConsumerOption configures the consumer.
type ConsumerOption func(*consumerOptions)

type consumerOptions struct {
    commitStrategy   CommitStrategy
    commitConfig     CommitConfig
    dlqHandler       DLQHandler
    interceptors     []ConsumerInterceptor
    errorHandler     ErrorHandler
    rebalanceHandler RebalanceHandler
    metricsEnabled   bool
}

// WithCommitStrategy sets the commit strategy.
func WithCommitStrategy(strategy CommitStrategy) ConsumerOption {
    return func(o *consumerOptions) {
        o.commitStrategy = strategy
    }
}

// WithDLQHandler sets the DLQ handler.
func WithDLQHandler(h DLQHandler) ConsumerOption {
    return func(o *consumerOptions) {
        o.dlqHandler = h
    }
}

// WithConsumerInterceptor adds an interceptor.
func WithConsumerInterceptor(i ConsumerInterceptor) ConsumerOption {
    return func(o *consumerOptions) {
        o.interceptors = append(o.interceptors, i)
    }
}

// WithRebalanceHandler sets the rebalance callback handler.
func WithRebalanceHandler(h RebalanceHandler) ConsumerOption {
    return func(o *consumerOptions) {
        o.rebalanceHandler = h
    }
}
```

### Consumer Implementation

```go
// pkg/kafka/consumer_impl.go

package kafka

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type consumer struct {
    config           *ConsumerConfig
    client           KafkaConsumerClient
    handler          MessageHandler
    dlqHandler       DLQHandler
    committer        Committer
    interceptors     []ConsumerInterceptor
    rebalanceHandler RebalanceHandler
    metrics          *ConsumerMetrics

    mu       sync.RWMutex
    running  bool
    closed   bool
    stopChan chan struct{}
    wg       sync.WaitGroup
}

// NewConsumer creates a new Kafka consumer with the given configuration.
func NewConsumer(cfg *Config, opts ...ConsumerOption) (Consumer, error) {
    options := &consumerOptions{
        commitStrategy: cfg.Consumer.CommitStrategy,
        commitConfig:   cfg.Consumer.CommitConfig,
    }

    for _, opt := range opts {
        opt(options)
    }

    client, err := newKafkaConsumerClient(cfg)
    if err != nil {
        return nil, fmt.Errorf("creating kafka consumer client: %w", err)
    }

    // Initialize DLQ handler if enabled
    var dlqHandler DLQHandler
    if cfg.DLQ.Enabled {
        dlqHandler, err = NewDLQHandler(cfg)
        if err != nil {
            return nil, fmt.Errorf("creating dlq handler: %w", err)
        }
    }
    if options.dlqHandler != nil {
        dlqHandler = options.dlqHandler
    }

    // Create committer based on strategy
    committer := newCommitter(client, options.commitStrategy, options.commitConfig)

    return &consumer{
        config:           &cfg.Consumer,
        client:           client,
        dlqHandler:       dlqHandler,
        committer:        committer,
        interceptors:     options.interceptors,
        rebalanceHandler: options.rebalanceHandler,
        metrics:          newConsumerMetrics(),
        stopChan:         make(chan struct{}),
    }, nil
}

func (c *consumer) Subscribe(topics ...string) error {
    return c.client.Subscribe(topics)
}

func (c *consumer) Start(ctx context.Context, handler MessageHandler) error {
    c.mu.Lock()
    if c.running {
        c.mu.Unlock()
        return ErrConsumerAlreadyRunning
    }
    c.running = true
    c.handler = handler
    c.mu.Unlock()

    // Start concurrent consumers if configured
    for i := 0; i < c.config.Concurrency; i++ {
        c.wg.Add(1)
        go c.consumeLoop(ctx, i)
    }

    // Start committer if using batch strategy
    if c.config.CommitStrategy == CommitStrategyBatch {
        c.wg.Add(1)
        go c.committer.Start(ctx, &c.wg, c.stopChan)
    }

    c.wg.Wait()
    return nil
}

func (c *consumer) consumeLoop(ctx context.Context, workerID int) {
    defer c.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case <-c.stopChan:
            return
        default:
            msg, err := c.client.Poll(ctx, c.config.MaxPollInterval)
            if err != nil {
                if err == ErrNoMessage {
                    continue
                }
                c.metrics.RecordError()
                continue
            }

            c.processMessage(ctx, msg)
        }
    }
}

func (c *consumer) processMessage(ctx context.Context, msg *Message) {
    startTime := time.Now()

    // Run pre-process interceptors
    for _, interceptor := range c.interceptors {
        var err error
        msg, err = interceptor.OnConsume(ctx, msg)
        if err != nil {
            c.handleError(ctx, msg, err)
            return
        }
    }

    // Process message with retry logic
    var lastErr error
    for attempt := 0; attempt <= c.dlqHandler.MaxRetries(); attempt++ {
        if attempt > 0 {
            backoff := c.calculateBackoff(attempt)
            select {
            case <-ctx.Done():
                return
            case <-time.After(backoff):
            }
        }

        if err := c.handler.Handle(ctx, msg); err != nil {
            lastErr = err
            c.metrics.RecordRetry(msg.Topic)
            continue
        }

        // Success - commit based on strategy
        c.handleCommit(ctx, msg)
        c.metrics.RecordSuccess(msg.Topic, time.Since(startTime))
        return
    }

    // All retries exhausted - send to DLQ
    if c.dlqHandler != nil {
        if err := c.dlqHandler.Send(ctx, msg, lastErr); err != nil {
            c.metrics.RecordDLQError(msg.Topic)
        } else {
            c.metrics.RecordDLQ(msg.Topic)
        }
    }

    // Commit even on DLQ to prevent reprocessing
    c.handleCommit(ctx, msg)
}

func (c *consumer) handleCommit(ctx context.Context, msg *Message) {
    switch c.config.CommitStrategy {
    case CommitStrategyAutoCommit:
        // Handled by Kafka client automatically
    case CommitStrategySync:
        _ = c.committer.CommitSync(ctx, msg)
    case CommitStrategyAsync:
        c.committer.CommitAsync(msg)
    case CommitStrategyBatch:
        c.committer.AddToBatch(msg)
    case CommitStrategyManual:
        // User handles via Commit() or CommitMessage()
    }
}

func (c *consumer) calculateBackoff(attempt int) time.Duration {
    if c.dlqHandler == nil {
        return time.Second
    }

    backoff := c.dlqHandler.RetryBackoff()
    multiplier := c.dlqHandler.BackoffMultiplier()
    maxBackoff := c.dlqHandler.MaxBackoff()

    for i := 1; i < attempt; i++ {
        backoff = time.Duration(float64(backoff) * multiplier)
        if backoff > maxBackoff {
            backoff = maxBackoff
            break
        }
    }

    return backoff
}

func (c *consumer) Commit(ctx context.Context) error {
    return c.committer.CommitSync(ctx, nil)
}

func (c *consumer) CommitMessage(ctx context.Context, msg *Message) error {
    return c.committer.CommitSync(ctx, msg)
}

func (c *consumer) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.closed {
        return nil
    }
    c.closed = true
    close(c.stopChan)

    if c.dlqHandler != nil {
        c.dlqHandler.Close()
    }

    return c.client.Close()
}

func (c *consumer) Lag() map[TopicPartition]int64 {
    return c.client.Lag()
}
```

---

## Dead Letter Queue (DLQ) Design

### DLQ Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Message Processing Flow                           │
└─────────────────────────────────────────────────────────────────────────────┘

    ┌──────────┐
    │  Main    │
    │  Topic   │
    └────┬─────┘
         │
         ▼
    ┌──────────┐     Success
    │ Consumer │─────────────────────────────────────────────┐
    │ Handler  │                                             │
    └────┬─────┘                                             │
         │ Failure                                           │
         ▼                                                   │
    ┌──────────┐                                             │
    │  Retry   │◄──────────────────────┐                     │
    │  Logic   │                       │                     │
    └────┬─────┘                       │                     │
         │                             │                     │
         ▼                             │                     │
    ┌──────────────┐    Retry < Max    │                     │
    │ Retry Count  │───────────────────┘                     │
    │    Check     │                                         │
    └──────┬───────┘                                         │
           │ Retry >= Max                                    │
           ▼                                                 │
    ┌──────────────┐     ┌─────────────┐                     │
    │ Retry Topic  │────►│   Delayed   │                     │
    │  (Optional)  │     │ Reprocessing│                     │
    └──────────────┘     └──────┬──────┘                     │
           │                    │                            │
           │ Still Failing      │ Success                    │
           ▼                    ▼                            ▼
    ┌──────────────┐     ┌─────────────────────────────────────┐
    │  DLQ Topic   │     │            Commit Offset            │
    │              │     │                                     │
    │  Headers:    │     └─────────────────────────────────────┘
    │  - error     │
    │  - retry_cnt │
    │  - orig_topic│
    │  - timestamp │
    └──────────────┘
```

### DLQ Interface

```go
// pkg/kafka/dlq.go

package kafka

import (
    "context"
    "time"
)

// DLQHandler manages dead letter queue operations.
type DLQHandler interface {
    // Send publishes a failed message to the DLQ topic.
    Send(ctx context.Context, msg *Message, err error) error

    // SendToRetry publishes a message to the retry topic for delayed reprocessing.
    SendToRetry(ctx context.Context, msg *Message, retryCount int) error

    // MaxRetries returns the configured maximum retry attempts.
    MaxRetries() int

    // RetryBackoff returns the base retry backoff duration.
    RetryBackoff() time.Duration

    // BackoffMultiplier returns the backoff multiplier for exponential backoff.
    BackoffMultiplier() float64

    // MaxBackoff returns the maximum backoff duration.
    MaxBackoff() time.Duration

    // Close releases resources.
    Close() error
}

// DLQMessage represents a message in the DLQ with additional metadata.
type DLQMessage struct {
    OriginalMessage *Message          `json:"original_message"`
    OriginalTopic   string            `json:"original_topic"`
    OriginalOffset  int64             `json:"original_offset"`
    Error           string            `json:"error"`
    ErrorType       string            `json:"error_type"`
    RetryCount      int               `json:"retry_count"`
    FirstFailedAt   time.Time         `json:"first_failed_at"`
    LastFailedAt    time.Time         `json:"last_failed_at"`
    ConsumerGroup   string            `json:"consumer_group"`
    Hostname        string            `json:"hostname"`
    Metadata        map[string]string `json:"metadata,omitempty"`
}
```

### DLQ Implementation

```go
// pkg/kafka/dlq_impl.go

package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strconv"
    "time"
)

// DLQ Header keys
const (
    HeaderDLQError         = "dlq-error"
    HeaderDLQErrorType     = "dlq-error-type"
    HeaderDLQRetryCount    = "dlq-retry-count"
    HeaderDLQOriginalTopic = "dlq-original-topic"
    HeaderDLQOriginalKey   = "dlq-original-key"
    HeaderDLQFirstFailed   = "dlq-first-failed-at"
    HeaderDLQLastFailed    = "dlq-last-failed-at"
    HeaderDLQConsumerGroup = "dlq-consumer-group"
    HeaderDLQHostname      = "dlq-hostname"
)

type dlqHandler struct {
    config        *DLQConfig
    producer      Producer
    consumerGroup string
    hostname      string
}

// NewDLQHandler creates a new DLQ handler.
func NewDLQHandler(cfg *Config) (DLQHandler, error) {
    producer, err := NewProducer(cfg)
    if err != nil {
        return nil, fmt.Errorf("creating dlq producer: %w", err)
    }

    hostname, _ := os.Hostname()

    return &dlqHandler{
        config:        &cfg.DLQ,
        producer:      producer,
        consumerGroup: cfg.Consumer.GroupID,
        hostname:      hostname,
    }, nil
}

func (d *dlqHandler) Send(ctx context.Context, msg *Message, err error) error {
    dlqTopic := msg.Topic + d.config.TopicSuffix

    // Extract retry count from headers
    retryCount := 0
    if msg.Headers != nil {
        if rc, ok := msg.Headers[HeaderDLQRetryCount]; ok {
            retryCount, _ = strconv.Atoi(rc)
        }
    }

    // Build DLQ message
    dlqMsg := &Message{
        Key:   msg.Key,
        Value: msg.Value,
        Headers: map[string]string{
            HeaderDLQError:         err.Error(),
            HeaderDLQErrorType:     fmt.Sprintf("%T", err),
            HeaderDLQRetryCount:    strconv.Itoa(retryCount),
            HeaderDLQOriginalTopic: msg.Topic,
            HeaderDLQLastFailed:    time.Now().UTC().Format(time.RFC3339),
            HeaderDLQConsumerGroup: d.consumerGroup,
            HeaderDLQHostname:      d.hostname,
        },
        Timestamp: time.Now(),
    }

    // Preserve original key if present
    if len(msg.Key) > 0 {
        dlqMsg.Headers[HeaderDLQOriginalKey] = string(msg.Key)
    }

    // Set first failed time
    if msg.Headers != nil {
        if firstFailed, ok := msg.Headers[HeaderDLQFirstFailed]; ok {
            dlqMsg.Headers[HeaderDLQFirstFailed] = firstFailed
        } else {
            dlqMsg.Headers[HeaderDLQFirstFailed] = time.Now().UTC().Format(time.RFC3339)
        }
    } else {
        dlqMsg.Headers[HeaderDLQFirstFailed] = time.Now().UTC().Format(time.RFC3339)
    }

    // Preserve original headers if configured
    if d.config.IncludeHeaders && msg.Headers != nil {
        for k, v := range msg.Headers {
            if _, exists := dlqMsg.Headers[k]; !exists {
                dlqMsg.Headers["orig-"+k] = v
            }
        }
    }

    return d.producer.Publish(ctx, dlqTopic, dlqMsg)
}

func (d *dlqHandler) SendToRetry(ctx context.Context, msg *Message, retryCount int) error {
    if !d.config.RetryTopicEnabled {
        return nil
    }

    retryTopic := msg.Topic + d.config.RetryTopicSuffix

    retryMsg := &Message{
        Key:   msg.Key,
        Value: msg.Value,
        Headers: map[string]string{
            HeaderDLQRetryCount:    strconv.Itoa(retryCount),
            HeaderDLQOriginalTopic: msg.Topic,
            HeaderDLQLastFailed:    time.Now().UTC().Format(time.RFC3339),
        },
        Timestamp: time.Now(),
    }

    // Preserve first failed time
    if msg.Headers != nil {
        if firstFailed, ok := msg.Headers[HeaderDLQFirstFailed]; ok {
            retryMsg.Headers[HeaderDLQFirstFailed] = firstFailed
        }
    }
    if _, ok := retryMsg.Headers[HeaderDLQFirstFailed]; !ok {
        retryMsg.Headers[HeaderDLQFirstFailed] = time.Now().UTC().Format(time.RFC3339)
    }

    return d.producer.Publish(ctx, retryTopic, retryMsg)
}

func (d *dlqHandler) MaxRetries() int {
    return d.config.MaxRetries
}

func (d *dlqHandler) RetryBackoff() time.Duration {
    return d.config.RetryBackoff
}

func (d *dlqHandler) BackoffMultiplier() float64 {
    return d.config.RetryBackoffMultiplier
}

func (d *dlqHandler) MaxBackoff() time.Duration {
    return d.config.RetryMaxBackoff
}

func (d *dlqHandler) Close() error {
    return d.producer.Close()
}
```

### DLQ Recovery Consumer

```go
// pkg/kafka/dlq_recovery.go

package kafka

import (
    "context"
    "fmt"
    "strconv"
    "time"
)

// DLQRecoveryHandler processes messages from DLQ for manual recovery.
type DLQRecoveryHandler interface {
    // Reprocess attempts to reprocess a DLQ message to the original topic.
    Reprocess(ctx context.Context, msg *Message) error

    // ReprocessBatch reprocesses multiple DLQ messages.
    ReprocessBatch(ctx context.Context, messages []*Message) error

    // ListDLQMessages retrieves messages from DLQ topic for inspection.
    ListDLQMessages(ctx context.Context, topic string, limit int) ([]*DLQMessage, error)
}

type dlqRecoveryHandler struct {
    producer Producer
    consumer Consumer
}

// NewDLQRecoveryHandler creates a handler for DLQ recovery operations.
func NewDLQRecoveryHandler(cfg *Config) (DLQRecoveryHandler, error) {
    producer, err := NewProducer(cfg)
    if err != nil {
        return nil, fmt.Errorf("creating recovery producer: %w", err)
    }

    // Create a separate consumer for DLQ reading
    dlqConsumerCfg := *cfg
    dlqConsumerCfg.Consumer.GroupID = cfg.Consumer.GroupID + "-dlq-recovery"
    dlqConsumerCfg.Consumer.AutoOffsetReset = "earliest"

    consumer, err := NewConsumer(&dlqConsumerCfg)
    if err != nil {
        producer.Close()
        return nil, fmt.Errorf("creating recovery consumer: %w", err)
    }

    return &dlqRecoveryHandler{
        producer: producer,
        consumer: consumer,
    }, nil
}

func (r *dlqRecoveryHandler) Reprocess(ctx context.Context, msg *Message) error {
    // Extract original topic from headers
    originalTopic, ok := msg.Headers[HeaderDLQOriginalTopic]
    if !ok {
        return fmt.Errorf("missing original topic header")
    }

    // Create recovery message
    recoveryMsg := &Message{
        Key:       msg.Key,
        Value:     msg.Value,
        Timestamp: time.Now(),
        Headers:   make(map[string]string),
    }

    // Preserve retry count for tracking
    if rc, ok := msg.Headers[HeaderDLQRetryCount]; ok {
        count, _ := strconv.Atoi(rc)
        recoveryMsg.Headers[HeaderDLQRetryCount] = strconv.Itoa(count + 1)
    }

    // Restore original key if present
    if originalKey, ok := msg.Headers[HeaderDLQOriginalKey]; ok {
        recoveryMsg.Key = []byte(originalKey)
    }

    return r.producer.Publish(ctx, originalTopic, recoveryMsg)
}

func (r *dlqRecoveryHandler) ReprocessBatch(ctx context.Context, messages []*Message) error {
    for _, msg := range messages {
        if err := r.Reprocess(ctx, msg); err != nil {
            return fmt.Errorf("reprocessing message: %w", err)
        }
    }
    return nil
}

func (r *dlqRecoveryHandler) ListDLQMessages(ctx context.Context, topic string, limit int) ([]*DLQMessage, error) {
    // Implementation would poll messages from DLQ topic without committing
    // This is for inspection purposes
    return nil, fmt.Errorf("not implemented")
}
```

---

## Partitioning Strategies

### Partitioner Interface

```go
// pkg/kafka/partitioner.go

package kafka

import (
    "hash"
    "hash/fnv"
    "math/rand"
    "sync"
    "sync/atomic"
)

// Partitioner determines which partition a message should be sent to.
type Partitioner interface {
    // Partition returns the partition number for a message.
    // Returns -1 to let Kafka decide (for round-robin scenarios).
    Partition(topic string, key, value []byte) (int32, error)

    // RequiresConsistency returns true if the same key should always go to the same partition.
    RequiresConsistency() bool
}

// PartitionerFactory creates partitioners.
type PartitionerFactory func() Partitioner

// newPartitioner creates a partitioner based on the configured type.
func newPartitioner(pType PartitionerType) Partitioner {
    switch pType {
    case PartitionerHash:
        return &HashPartitioner{hasher: fnv.New32a()}
    case PartitionerRoundRobin:
        return &RoundRobinPartitioner{}
    case PartitionerRandom:
        return &RandomPartitioner{}
    case PartitionerSticky:
        return &StickyPartitioner{}
    default:
        return &HashPartitioner{hasher: fnv.New32a()}
    }
}

// HashPartitioner distributes messages based on key hash.
type HashPartitioner struct {
    hasher hash.Hash32
    mu     sync.Mutex
}

func (p *HashPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    if len(key) == 0 {
        // No key - return -1 to let Kafka handle it
        return -1, nil
    }

    p.mu.Lock()
    defer p.mu.Unlock()

    p.hasher.Reset()
    p.hasher.Write(key)
    hash := p.hasher.Sum32()

    // Return hash value; actual partition will be hash % numPartitions
    // The kafka client handles the modulo operation
    return int32(hash), nil
}

func (p *HashPartitioner) RequiresConsistency() bool {
    return true
}

// RoundRobinPartitioner distributes messages evenly across partitions.
type RoundRobinPartitioner struct {
    counter uint64
}

func (p *RoundRobinPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    // Return -1 to use Kafka's built-in round-robin
    return -1, nil
}

func (p *RoundRobinPartitioner) RequiresConsistency() bool {
    return false
}

// RandomPartitioner distributes messages randomly.
type RandomPartitioner struct{}

func (p *RandomPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    return int32(rand.Int31()), nil
}

func (p *RandomPartitioner) RequiresConsistency() bool {
    return false
}

// StickyPartitioner batches messages to the same partition until batch is full.
type StickyPartitioner struct {
    currentPartition int32
    messageCount     uint64
    batchSize        uint64
    mu               sync.RWMutex
}

func NewStickyPartitioner(batchSize uint64) *StickyPartitioner {
    return &StickyPartitioner{
        currentPartition: -1,
        batchSize:        batchSize,
    }
}

func (p *StickyPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    count := atomic.AddUint64(&p.messageCount, 1)

    if count%p.batchSize == 0 {
        p.mu.Lock()
        p.currentPartition = int32(rand.Int31())
        p.mu.Unlock()
    }

    p.mu.RLock()
    partition := p.currentPartition
    p.mu.RUnlock()

    return partition, nil
}

func (p *StickyPartitioner) RequiresConsistency() bool {
    return false
}

// KeyFieldPartitioner extracts a field from JSON messages for partitioning.
type KeyFieldPartitioner struct {
    fieldPath string
    hasher    hash.Hash32
    mu        sync.Mutex
}

func NewKeyFieldPartitioner(fieldPath string) *KeyFieldPartitioner {
    return &KeyFieldPartitioner{
        fieldPath: fieldPath,
        hasher:    fnv.New32a(),
    }
}

func (p *KeyFieldPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    // Extract field value from JSON
    fieldValue, err := extractJSONField(value, p.fieldPath)
    if err != nil {
        return -1, nil // Fallback to Kafka default
    }

    p.mu.Lock()
    defer p.mu.Unlock()

    p.hasher.Reset()
    p.hasher.Write(fieldValue)
    return int32(p.hasher.Sum32()), nil
}

func (p *KeyFieldPartitioner) RequiresConsistency() bool {
    return true
}
```

---

## Commit Strategies

### Committer Interface

```go
// pkg/kafka/committer.go

package kafka

import (
    "context"
    "sync"
    "time"
)

// Committer handles offset commits based on the configured strategy.
type Committer interface {
    // CommitSync synchronously commits offsets.
    CommitSync(ctx context.Context, msg *Message) error

    // CommitAsync asynchronously commits offsets.
    CommitAsync(msg *Message)

    // AddToBatch adds a message to the commit batch.
    AddToBatch(msg *Message)

    // Start begins the commit loop (for batch strategy).
    Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{})

    // Flush commits all pending offsets.
    Flush(ctx context.Context) error
}

// newCommitter creates a committer based on the strategy.
func newCommitter(client KafkaConsumerClient, strategy CommitStrategy, config CommitConfig) Committer {
    switch strategy {
    case CommitStrategyAutoCommit:
        return &autoCommitter{client: client}
    case CommitStrategySync:
        return &syncCommitter{client: client}
    case CommitStrategyAsync:
        return &asyncCommitter{client: client}
    case CommitStrategyBatch:
        return &batchCommitter{
            client:    client,
            batchSize: config.BatchCommitSize,
            interval:  config.BatchCommitInterval,
            batch:     make([]*Message, 0, config.BatchCommitSize),
        }
    case CommitStrategyManual:
        return &manualCommitter{client: client}
    default:
        return &autoCommitter{client: client}
    }
}

// autoCommitter relies on Kafka's auto-commit feature.
type autoCommitter struct {
    client KafkaConsumerClient
}

func (c *autoCommitter) CommitSync(ctx context.Context, msg *Message) error {
    return nil // Auto-commit handles it
}

func (c *autoCommitter) CommitAsync(msg *Message) {}

func (c *autoCommitter) AddToBatch(msg *Message) {}

func (c *autoCommitter) Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()
}

func (c *autoCommitter) Flush(ctx context.Context) error {
    return nil
}

// syncCommitter commits offsets synchronously after each message.
type syncCommitter struct {
    client KafkaConsumerClient
}

func (c *syncCommitter) CommitSync(ctx context.Context, msg *Message) error {
    if msg == nil {
        return c.client.Commit(ctx)
    }
    return c.client.CommitMessage(ctx, msg)
}

func (c *syncCommitter) CommitAsync(msg *Message) {
    // Not supported in sync mode
}

func (c *syncCommitter) AddToBatch(msg *Message) {}

func (c *syncCommitter) Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()
}

func (c *syncCommitter) Flush(ctx context.Context) error {
    return c.client.Commit(ctx)
}

// asyncCommitter commits offsets asynchronously.
type asyncCommitter struct {
    client KafkaConsumerClient
}

func (c *asyncCommitter) CommitSync(ctx context.Context, msg *Message) error {
    return c.client.CommitMessage(ctx, msg)
}

func (c *asyncCommitter) CommitAsync(msg *Message) {
    go func() {
        _ = c.client.CommitMessage(context.Background(), msg)
    }()
}

func (c *asyncCommitter) AddToBatch(msg *Message) {}

func (c *asyncCommitter) Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()
}

func (c *asyncCommitter) Flush(ctx context.Context) error {
    return c.client.Commit(ctx)
}

// batchCommitter accumulates messages and commits in batches.
type batchCommitter struct {
    client    KafkaConsumerClient
    batchSize int
    interval  time.Duration

    mu    sync.Mutex
    batch []*Message
}

func (c *batchCommitter) CommitSync(ctx context.Context, msg *Message) error {
    return c.client.CommitMessage(ctx, msg)
}

func (c *batchCommitter) CommitAsync(msg *Message) {
    c.AddToBatch(msg)
}

func (c *batchCommitter) AddToBatch(msg *Message) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.batch = append(c.batch, msg)
}

func (c *batchCommitter) Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()

    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            c.flush(ctx)
            return
        case <-stopChan:
            c.flush(ctx)
            return
        case <-ticker.C:
            c.flush(ctx)
        }
    }
}

func (c *batchCommitter) flush(ctx context.Context) {
    c.mu.Lock()
    batch := c.batch
    c.batch = make([]*Message, 0, c.batchSize)
    c.mu.Unlock()

    if len(batch) == 0 {
        return
    }

    // Commit highest offset per partition
    offsets := make(map[TopicPartition]*Message)
    for _, msg := range batch {
        tp := TopicPartition{Topic: msg.Topic, Partition: msg.Partition}
        if existing, ok := offsets[tp]; !ok || msg.Offset > existing.Offset {
            offsets[tp] = msg
        }
    }

    for _, msg := range offsets {
        _ = c.client.CommitMessage(ctx, msg)
    }
}

func (c *batchCommitter) Flush(ctx context.Context) error {
    c.flush(ctx)
    return nil
}

// manualCommitter leaves commit control to the user.
type manualCommitter struct {
    client KafkaConsumerClient
}

func (c *manualCommitter) CommitSync(ctx context.Context, msg *Message) error {
    if msg == nil {
        return c.client.Commit(ctx)
    }
    return c.client.CommitMessage(ctx, msg)
}

func (c *manualCommitter) CommitAsync(msg *Message) {}

func (c *manualCommitter) AddToBatch(msg *Message) {}

func (c *manualCommitter) Start(ctx context.Context, wg *sync.WaitGroup, stopChan <-chan struct{}) {
    defer wg.Done()
}

func (c *manualCommitter) Flush(ctx context.Context) error {
    return c.client.Commit(ctx)
}
```

---

## Error Handling

### Error Types

```go
// pkg/kafka/errors.go

package kafka

import (
    "errors"
    "fmt"
)

// Sentinel errors
var (
    ErrProducerClosed         = errors.New("kafka: producer is closed")
    ErrConsumerClosed         = errors.New("kafka: consumer is closed")
    ErrConsumerAlreadyRunning = errors.New("kafka: consumer is already running")
    ErrNoMessage              = errors.New("kafka: no message available")
    ErrInvalidConfig          = errors.New("kafka: invalid configuration")
    ErrSerializationFailed    = errors.New("kafka: serialization failed")
    ErrDeserializationFailed  = errors.New("kafka: deserialization failed")
    ErrPartitioningFailed     = errors.New("kafka: partitioning failed")
    ErrCommitFailed           = errors.New("kafka: commit failed")
    ErrDLQFailed              = errors.New("kafka: dlq send failed")
)

// RetryableError indicates the error is transient and can be retried.
type RetryableError struct {
    Err error
}

func (e *RetryableError) Error() string {
    return fmt.Sprintf("retryable: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
    return e.Err
}

// IsRetryable checks if an error should be retried.
func IsRetryable(err error) bool {
    var retryable *RetryableError
    return errors.As(err, &retryable)
}

// PermanentError indicates the error is not recoverable.
type PermanentError struct {
    Err error
}

func (e *PermanentError) Error() string {
    return fmt.Sprintf("permanent: %v", e.Err)
}

func (e *PermanentError) Unwrap() error {
    return e.Err
}

// IsPermanent checks if an error should not be retried.
func IsPermanent(err error) bool {
    var permanent *PermanentError
    return errors.As(err, &permanent)
}

// BatchError contains errors from batch operations.
type BatchError struct {
    Errors []error
}

func (e *BatchError) Error() string {
    return fmt.Sprintf("batch operation failed with %d errors", len(e.Errors))
}

// ErrorHandler handles errors during message processing.
type ErrorHandler interface {
    OnError(ctx context.Context, err error, msg *Message)
}

// ErrorHandlerFunc is a function adapter for ErrorHandler.
type ErrorHandlerFunc func(ctx context.Context, err error, msg *Message)

func (f ErrorHandlerFunc) OnError(ctx context.Context, err error, msg *Message) {
    f(ctx, err, msg)
}
```

---

## Interfaces & Models

### Supporting Interfaces

```go
// pkg/kafka/interfaces.go

package kafka

import (
    "context"
)

// KafkaClient abstracts the underlying Kafka client library.
type KafkaClient interface {
    Produce(ctx context.Context, topic string, partition int32, msg *Message) error
    Flush(ctx context.Context) error
    Close() error
}

// KafkaConsumerClient abstracts consumer operations.
type KafkaConsumerClient interface {
    Subscribe(topics []string) error
    Poll(ctx context.Context, timeout time.Duration) (*Message, error)
    Commit(ctx context.Context) error
    CommitMessage(ctx context.Context, msg *Message) error
    Pause(partitions []TopicPartition) error
    Resume(partitions []TopicPartition) error
    Lag() map[TopicPartition]int64
    Close() error
}

// Serializer handles message serialization.
type Serializer interface {
    Serialize(v interface{}) ([]byte, error)
    ContentType() string
}

// Deserializer handles message deserialization.
type Deserializer interface {
    Deserialize(data []byte, v interface{}) error
}

// ProducerInterceptor intercepts messages before sending.
type ProducerInterceptor interface {
    OnSend(ctx context.Context, topic string, msg *Message) (*Message, error)
}

// ConsumerInterceptor intercepts messages after receiving.
type ConsumerInterceptor interface {
    OnConsume(ctx context.Context, msg *Message) (*Message, error)
}

// RebalanceHandler handles consumer group rebalancing events.
type RebalanceHandler interface {
    OnPartitionsAssigned(partitions []TopicPartition)
    OnPartitionsRevoked(partitions []TopicPartition)
}

// JSONSerializer implements JSON serialization.
type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

func (s *JSONSerializer) ContentType() string {
    return "application/json"
}

// JSONDeserializer implements JSON deserialization.
type JSONDeserializer struct{}

func (d *JSONDeserializer) Deserialize(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}
```

### Metrics

```go
// pkg/kafka/metrics.go

package kafka

import (
    "sync"
    "sync/atomic"
    "time"
)

// ProducerMetrics tracks producer statistics.
type ProducerMetrics struct {
    messagesSent    uint64
    messagesErrored uint64
    bytesSent       uint64

    topicMetrics sync.Map // map[string]*TopicMetrics
}

// ConsumerMetrics tracks consumer statistics.
type ConsumerMetrics struct {
    messagesReceived  uint64
    messagesProcessed uint64
    messagesErrored   uint64
    messagesRetried   uint64
    messagesDLQ       uint64
    dlqErrors         uint64

    processingTime sync.Map // map[string]*LatencyMetrics
}

// TopicMetrics tracks per-topic statistics.
type TopicMetrics struct {
    sent    uint64
    errored uint64
}

// LatencyMetrics tracks processing latency.
type LatencyMetrics struct {
    count      uint64
    totalNanos uint64
    minNanos   uint64
    maxNanos   uint64
}

func newProducerMetrics() *ProducerMetrics {
    return &ProducerMetrics{}
}

func (m *ProducerMetrics) RecordSuccess(topic string) {
    atomic.AddUint64(&m.messagesSent, 1)
}

func (m *ProducerMetrics) RecordError(topic string) {
    atomic.AddUint64(&m.messagesErrored, 1)
}

func newConsumerMetrics() *ConsumerMetrics {
    return &ConsumerMetrics{}
}

func (m *ConsumerMetrics) RecordSuccess(topic string, duration time.Duration) {
    atomic.AddUint64(&m.messagesProcessed, 1)
}

func (m *ConsumerMetrics) RecordError() {
    atomic.AddUint64(&m.messagesErrored, 1)
}

func (m *ConsumerMetrics) RecordRetry(topic string) {
    atomic.AddUint64(&m.messagesRetried, 1)
}

func (m *ConsumerMetrics) RecordDLQ(topic string) {
    atomic.AddUint64(&m.messagesDLQ, 1)
}

func (m *ConsumerMetrics) RecordDLQError(topic string) {
    atomic.AddUint64(&m.dlqErrors, 1)
}
```

---

## Package Structure

```
pkg/
└── kafka/
    ├── config.go           # Configuration structs and loading
    ├── config_test.go      # Configuration tests
    ├── producer.go         # Producer interface
    ├── producer_impl.go    # Producer implementation
    ├── producer_test.go    # Producer tests
    ├── consumer.go         # Consumer interface
    ├── consumer_impl.go    # Consumer implementation
    ├── consumer_test.go    # Consumer tests
    ├── partitioner.go      # Partitioning strategies
    ├── partitioner_test.go # Partitioner tests
    ├── committer.go        # Commit strategies
    ├── committer_test.go   # Committer tests
    ├── dlq.go              # DLQ interface
    ├── dlq_impl.go         # DLQ implementation
    ├── dlq_recovery.go     # DLQ recovery tools
    ├── dlq_test.go         # DLQ tests
    ├── errors.go           # Error types
    ├── interfaces.go       # Common interfaces
    ├── metrics.go          # Metrics collection
    ├── serializer.go       # Serialization helpers
    ├── client_confluent.go # Confluent Kafka client adapter
    ├── client_sarama.go    # Sarama client adapter (optional)
    └── testutil/
        └── mocks.go        # Test mocks
```

---

## Usage Examples

### Basic Producer

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/kafka"
)

func main() {
    cfg, err := kafka.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    producer, err := kafka.NewProducer(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer producer.Close()

    ctx := context.Background()

    msg := &kafka.Message{
        Key:   []byte("user-123"),
        Value: []byte(`{"event": "user_created", "user_id": "123"}`),
        Headers: map[string]string{
            "content-type": "application/json",
            "trace-id":     "abc-123",
        },
    }

    if err := producer.Publish(ctx, "user-events", msg); err != nil {
        log.Printf("Failed to publish: %v", err)
    }
}
```

### Basic Consumer with DLQ

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/rompi/core-backend/pkg/kafka"
)

func main() {
    cfg, err := kafka.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Enable DLQ
    cfg.DLQ.Enabled = true
    cfg.DLQ.MaxRetries = 3

    consumer, err := kafka.NewConsumer(cfg,
        kafka.WithCommitStrategy(kafka.CommitStrategySync),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := consumer.Subscribe("user-events"); err != nil {
        log.Fatal(err)
    }

    // Handle shutdown
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // Process messages
    handler := kafka.MessageHandlerFunc(func(ctx context.Context, msg *kafka.Message) error {
        log.Printf("Received message: %s", string(msg.Value))

        // Process message...
        // Return error to trigger retry/DLQ
        return nil
    })

    if err := consumer.Start(ctx, handler); err != nil {
        log.Printf("Consumer stopped: %v", err)
    }

    consumer.Close()
}
```

### Custom Partitioner

```go
package main

import (
    "github.com/rompi/core-backend/pkg/kafka"
)

// TenantPartitioner routes messages based on tenant ID.
type TenantPartitioner struct {
    hasher hash.Hash32
}

func (p *TenantPartitioner) Partition(topic string, key, value []byte) (int32, error) {
    // Extract tenant ID from message value
    var msg struct {
        TenantID string `json:"tenant_id"`
    }

    if err := json.Unmarshal(value, &msg); err != nil {
        return -1, nil // Fallback to default
    }

    p.hasher.Reset()
    p.hasher.Write([]byte(msg.TenantID))
    return int32(p.hasher.Sum32()), nil
}

func (p *TenantPartitioner) RequiresConsistency() bool {
    return true
}

func main() {
    cfg, _ := kafka.LoadConfig()

    producer, _ := kafka.NewProducer(cfg,
        kafka.WithPartitioner(&TenantPartitioner{hasher: fnv.New32a()}),
    )
    defer producer.Close()

    // Messages with same tenant_id go to same partition
}
```

### Batch Commit Strategy

```go
package main

import (
    "github.com/rompi/core-backend/pkg/kafka"
)

func main() {
    cfg, _ := kafka.LoadConfig()

    // Configure batch commits
    cfg.Consumer.CommitStrategy = kafka.CommitStrategyBatch
    cfg.Consumer.CommitConfig.BatchCommitSize = 100
    cfg.Consumer.CommitConfig.BatchCommitInterval = 10 * time.Second

    consumer, _ := kafka.NewConsumer(cfg)

    // Consumer will batch commits for efficiency
    // Commits happen every 100 messages or 10 seconds, whichever comes first
}
```

### DLQ Recovery

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/kafka"
)

func main() {
    cfg, _ := kafka.LoadConfig()

    recovery, err := kafka.NewDLQRecoveryHandler(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List DLQ messages for inspection
    messages, _ := recovery.ListDLQMessages(ctx, "user-events.dlq", 100)

    for _, msg := range messages {
        log.Printf("DLQ Message: topic=%s, error=%s, retries=%d",
            msg.OriginalTopic, msg.Error, msg.RetryCount)
    }

    // Reprocess selected messages
    for _, msg := range messages {
        if shouldReprocess(msg) {
            if err := recovery.Reprocess(ctx, msg.OriginalMessage); err != nil {
                log.Printf("Reprocess failed: %v", err)
            }
        }
    }
}
```

---

## Configuration Comparison: Commit Strategies

| Strategy | Description | Use Case | Trade-offs |
|----------|-------------|----------|------------|
| `auto_commit` | Kafka auto-commits at intervals | High throughput, at-least-once OK | May reprocess on crash |
| `sync` | Commit after each message | Strong consistency needed | Lower throughput |
| `async` | Commit asynchronously | Balance of throughput and safety | Some reprocessing possible |
| `batch` | Commit batches periodically | High throughput, some latency OK | Batch reprocessing on failure |
| `manual` | Application controls commits | Complex processing logic | Requires careful management |

## Configuration Comparison: Partitioners

| Partitioner | Description | Use Case | Ordering Guarantee |
|-------------|-------------|----------|-------------------|
| `hash` | Hash-based on key | Same key to same partition | Per-key ordering |
| `round_robin` | Even distribution | Maximum parallelism | No ordering |
| `random` | Random distribution | Testing, no ordering needs | No ordering |
| `sticky` | Batch to same partition | Reduce broker requests | Within batch only |
| `custom` | User-defined logic | Domain-specific routing | Depends on impl |

---

## Next Steps

1. **Implementation**: Implement the design using `confluent-kafka-go` or `sarama`
2. **Testing**: Add comprehensive unit and integration tests
3. **Monitoring**: Integrate with Prometheus/OpenTelemetry for observability
4. **Documentation**: Add GoDoc comments and usage examples
5. **CLI Tools**: Build CLI for DLQ inspection and recovery

---

## Dependencies to Add

```go
// go.mod additions
require (
    github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
    // OR
    github.com/IBM/sarama v1.42.0
)
```
