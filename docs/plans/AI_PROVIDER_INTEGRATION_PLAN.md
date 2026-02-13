# AI Provider Integration Plan

## Overview

This plan outlines the design and implementation of a **provider-agnostic AI integration package** (`pkg/aiclient`) for the core-backend library. The package will support multiple AI providers (OpenAI, Anthropic, Google AI, Azure OpenAI, etc.) through a unified interface while following the architectural patterns established in this codebase.

## Goals

1. **Provider Agnostic**: Single interface that works with any AI provider
2. **Modular Design**: Follow existing patterns (Repository, Middleware, Builder)
3. **Production Ready**: Include retry logic, circuit breaker, rate limiting
4. **Streaming Support**: Support both streaming and non-streaming responses
5. **Testable**: Mock implementations for testing
6. **Observable**: Built-in logging and metrics hooks
7. **Zero Lock-in**: Easy to switch providers or use multiple simultaneously

## Architecture

### Core Design Patterns

Following the existing codebase patterns:

1. **Interface-Based Abstraction** (like `auth.Service`)
2. **Repository Pattern** (like `auth` repositories)
3. **Middleware/Interceptor Chain** (like `httpclient.Middleware`)
4. **Builder/Fluent API** (like `httpclient.RequestBuilder`)
5. **Functional Options** (like `server.Option`)
6. **Dependency Injection** (all packages)

### Package Structure

```
pkg/aiclient/
├── client.go              # Main Client interface
├── provider.go            # Provider interface abstraction
├── request.go             # Request builder (fluent API)
├── response.go            # Response models
├── config.go              # Configuration structs
├── errors.go              # Error definitions
├── logger.go              # Logger interface
├── middleware.go          # Middleware chain support
├── retry.go               # Retry logic with backoff
├── circuitbreaker.go      # Circuit breaker pattern
├── ratelimiter.go         # Rate limiting
├── streaming.go           # Streaming support
├── models.go              # Domain models (Message, etc.)
├── providers/
│   ├── provider.go        # Base provider interface
│   ├── openai/
│   │   ├── provider.go
│   │   ├── models.go
│   │   └── streaming.go
│   ├── anthropic/
│   │   ├── provider.go
│   │   ├── models.go
│   │   └── streaming.go
│   ├── google/
│   │   ├── provider.go
│   │   ├── models.go
│   │   └── streaming.go
│   └── azure/
│       ├── provider.go
│       ├── models.go
│       └── streaming.go
├── testutil/
│   ├── mock_provider.go   # Mock provider for testing
│   └── fixtures.go        # Test data fixtures
├── examples/
│   ├── basic/
│   ├── streaming/
│   ├── multi-provider/
│   ├── with-middleware/
│   └── with-server/
└── README.md

```

## Core Interfaces

### 1. Client Interface

The main interface that consumers interact with:

```go
package aiclient

import (
    "context"
    "io"
)

// Client is the main interface for interacting with AI providers
type Client interface {
    // Chat sends a chat completion request
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream sends a chat completion request with streaming response
    ChatStream(ctx context.Context, req *ChatRequest) (ChatStream, error)

    // Embeddings generates embeddings for input text
    Embeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

    // Close closes any underlying connections
    Close() error
}

// ChatStream represents a streaming chat response
type ChatStream interface {
    // Recv receives the next chunk from the stream
    Recv() (*ChatStreamChunk, error)

    // Close closes the stream
    Close() error
}
```

### 2. Provider Interface

Abstract interface that all providers must implement:

```go
// Provider is the interface that all AI providers must implement
type Provider interface {
    // Name returns the provider name (e.g., "openai", "anthropic")
    Name() string

    // Chat sends a chat completion request to the provider
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream sends a streaming chat completion request
    ChatStream(ctx context.Context, req *ChatRequest) (ChatStream, error)

    // Embeddings generates embeddings
    Embeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

    // ValidateConfig validates the provider configuration
    ValidateConfig() error

    // Close closes any underlying connections
    Close() error
}
```

### 3. Request Builder (Fluent API)

Similar to `httpclient.RequestBuilder`:

```go
// RequestBuilder provides a fluent API for building chat requests
type RequestBuilder struct {
    client   Client
    messages []Message
    opts     *ChatOptions
    err      error
}

// NewRequest creates a new request builder
func (c *client) NewRequest() *RequestBuilder {
    return &RequestBuilder{
        client:   c,
        messages: make([]Message, 0),
        opts:     &ChatOptions{},
    }
}

// System adds a system message
func (rb *RequestBuilder) System(content string) *RequestBuilder {
    rb.messages = append(rb.messages, Message{
        Role:    RoleSystem,
        Content: content,
    })
    return rb
}

// User adds a user message
func (rb *RequestBuilder) User(content string) *RequestBuilder {
    rb.messages = append(rb.messages, Message{
        Role:    RoleUser,
        Content: content,
    })
    return rb
}

// Assistant adds an assistant message
func (rb *RequestBuilder) Assistant(content string) *RequestBuilder {
    rb.messages = append(rb.messages, Message{
        Role:    RoleAssistant,
        Content: content,
    })
    return rb
}

// Model sets the model to use
func (rb *RequestBuilder) Model(model string) *RequestBuilder {
    rb.opts.Model = model
    return rb
}

// Temperature sets the temperature
func (rb *RequestBuilder) Temperature(temp float64) *RequestBuilder {
    rb.opts.Temperature = &temp
    return rb
}

// MaxTokens sets the max tokens
func (rb *RequestBuilder) MaxTokens(max int) *RequestBuilder {
    rb.opts.MaxTokens = &max
    return rb
}

// Do executes the request
func (rb *RequestBuilder) Do(ctx context.Context) (*ChatResponse, error) {
    if rb.err != nil {
        return nil, rb.err
    }

    req := &ChatRequest{
        Messages: rb.messages,
        Options:  rb.opts,
    }

    return rb.client.Chat(ctx, req)
}

// Stream executes the request with streaming
func (rb *RequestBuilder) Stream(ctx context.Context) (ChatStream, error) {
    if rb.err != nil {
        return nil, rb.err
    }

    req := &ChatRequest{
        Messages: rb.messages,
        Options:  rb.opts,
    }

    return rb.client.ChatStream(ctx, req)
}
```

### 4. Middleware Interface

Similar to `httpclient.Middleware`:

```go
// Middleware is a function that wraps a Provider
type Middleware func(Provider) Provider

// Use adds middleware to the client
func (c *client) Use(middleware ...Middleware) {
    for _, m := range middleware {
        c.provider = m(c.provider)
    }
}
```

## Domain Models

### Universal Message Format

Provider-agnostic message structure:

```go
// Message represents a chat message
type Message struct {
    Role       MessageRole       `json:"role"`
    Content    string            `json:"content"`
    Name       string            `json:"name,omitempty"`
    ToolCalls  []ToolCall        `json:"tool_calls,omitempty"`
    ToolCallID string            `json:"tool_call_id,omitempty"`
    Metadata   map[string]string `json:"metadata,omitempty"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
    RoleSystem    MessageRole = "system"
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleTool      MessageRole = "tool"
)

// ChatRequest represents a chat completion request
type ChatRequest struct {
    Messages []Message    `json:"messages"`
    Options  *ChatOptions `json:"options,omitempty"`
}

// ChatOptions contains optional parameters for chat requests
type ChatOptions struct {
    Model            string              `json:"model,omitempty"`
    Temperature      *float64            `json:"temperature,omitempty"`
    MaxTokens        *int                `json:"max_tokens,omitempty"`
    TopP             *float64            `json:"top_p,omitempty"`
    FrequencyPenalty *float64            `json:"frequency_penalty,omitempty"`
    PresencePenalty  *float64            `json:"presence_penalty,omitempty"`
    Stop             []string            `json:"stop,omitempty"`
    Tools            []Tool              `json:"tools,omitempty"`
    ToolChoice       *ToolChoice         `json:"tool_choice,omitempty"`
    ResponseFormat   *ResponseFormat     `json:"response_format,omitempty"`
    User             string              `json:"user,omitempty"`
    Metadata         map[string]string   `json:"metadata,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
    ID                string         `json:"id"`
    Model             string         `json:"model"`
    Message           Message        `json:"message"`
    FinishReason      string         `json:"finish_reason"`
    Usage             Usage          `json:"usage"`
    CreatedAt         int64          `json:"created_at"`
    Provider          string         `json:"provider"`
    RawResponse       interface{}    `json:"-"` // Original provider response
}

// Usage represents token usage information
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// Tool represents a function tool
type Tool struct {
    Type     string       `json:"type"`
    Function ToolFunction `json:"function"`
}

// ToolFunction represents a function definition
type ToolFunction struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call made by the assistant
type ToolCall struct {
    ID       string   `json:"id"`
    Type     string   `json:"type"`
    Function Function `json:"function"`
}

// Function represents a function call
type Function struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

## Configuration

### Config Structure

```go
// Config represents the AI client configuration
type Config struct {
    // Provider is the AI provider to use
    Provider string `json:"provider"`

    // APIKey is the API key for authentication
    APIKey string `json:"api_key"`

    // BaseURL is the base URL for the API (optional, provider-specific)
    BaseURL string `json:"base_url,omitempty"`

    // DefaultModel is the default model to use
    DefaultModel string `json:"default_model,omitempty"`

    // Timeout is the request timeout
    Timeout time.Duration `json:"timeout"`

    // MaxRetries is the maximum number of retries
    MaxRetries int `json:"max_retries"`

    // RetryBackoff is the initial retry backoff duration
    RetryBackoff time.Duration `json:"retry_backoff"`

    // CircuitBreaker enables/disables circuit breaker
    CircuitBreaker bool `json:"circuit_breaker"`

    // CircuitBreakerThreshold is the failure threshold
    CircuitBreakerThreshold int `json:"circuit_breaker_threshold"`

    // RateLimit is the rate limit (requests per second)
    RateLimit float64 `json:"rate_limit,omitempty"`

    // Logger is the logger instance
    Logger Logger `json:"-"`

    // HTTPClient is the underlying HTTP client
    HTTPClient *http.Client `json:"-"`
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
    cfg := &Config{
        Provider:                 getEnv("AI_PROVIDER", "openai"),
        APIKey:                   getEnv("AI_API_KEY", ""),
        BaseURL:                  getEnv("AI_BASE_URL", ""),
        DefaultModel:             getEnv("AI_DEFAULT_MODEL", ""),
        Timeout:                  parseDuration(getEnv("AI_TIMEOUT", "30s")),
        MaxRetries:               parseInt(getEnv("AI_MAX_RETRIES", "3")),
        RetryBackoff:             parseDuration(getEnv("AI_RETRY_BACKOFF", "1s")),
        CircuitBreaker:           parseBool(getEnv("AI_CIRCUIT_BREAKER", "true")),
        CircuitBreakerThreshold:  parseInt(getEnv("AI_CIRCUIT_BREAKER_THRESHOLD", "5")),
        RateLimit:                parseFloat(getEnv("AI_RATE_LIMIT", "0")),
    }

    return cfg, cfg.Validate()
}

// Validate validates the configuration
func (c *Config) Validate() error {
    if c.Provider == "" {
        return ErrInvalidProvider
    }

    if c.APIKey == "" {
        return ErrMissingAPIKey
    }

    return nil
}
```

### Environment Variables

```bash
# Provider configuration
AI_PROVIDER=openai                    # Provider: openai, anthropic, google, azure
AI_API_KEY=sk-...                     # API key
AI_BASE_URL=                          # Optional base URL override
AI_DEFAULT_MODEL=gpt-4                # Default model

# Request configuration
AI_TIMEOUT=30s                        # Request timeout
AI_MAX_RETRIES=3                      # Max retry attempts
AI_RETRY_BACKOFF=1s                   # Initial backoff duration

# Resilience configuration
AI_CIRCUIT_BREAKER=true               # Enable circuit breaker
AI_CIRCUIT_BREAKER_THRESHOLD=5        # Failure threshold
AI_RATE_LIMIT=0                       # Rate limit (0 = disabled)
```

## Provider Implementations

### OpenAI Provider

```go
package openai

import (
    "context"
    "github.com/yourusername/core-backend/pkg/aiclient"
)

// Provider implements the OpenAI provider
type Provider struct {
    config     *Config
    httpClient *http.Client
    logger     aiclient.Logger
}

// Config represents OpenAI-specific configuration
type Config struct {
    APIKey       string
    BaseURL      string // defaults to https://api.openai.com/v1
    Organization string
    DefaultModel string
}

// NewProvider creates a new OpenAI provider
func NewProvider(cfg *Config, opts ...Option) (*Provider, error) {
    if cfg.BaseURL == "" {
        cfg.BaseURL = "https://api.openai.com/v1"
    }

    if cfg.DefaultModel == "" {
        cfg.DefaultModel = "gpt-4-turbo-preview"
    }

    p := &Provider{
        config:     cfg,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }

    for _, opt := range opts {
        opt(p)
    }

    return p, p.ValidateConfig()
}

// Name returns the provider name
func (p *Provider) Name() string {
    return "openai"
}

// Chat sends a chat completion request
func (p *Provider) Chat(ctx context.Context, req *aiclient.ChatRequest) (*aiclient.ChatResponse, error) {
    // Convert universal request to OpenAI format
    openaiReq := p.toOpenAIRequest(req)

    // Make HTTP request
    resp, err := p.doRequest(ctx, "/chat/completions", openaiReq)
    if err != nil {
        return nil, err
    }

    // Convert OpenAI response to universal format
    return p.fromOpenAIResponse(resp)
}

// ChatStream sends a streaming chat completion request
func (p *Provider) ChatStream(ctx context.Context, req *aiclient.ChatRequest) (aiclient.ChatStream, error) {
    // Implementation with SSE parsing
}
```

### Anthropic Provider

```go
package anthropic

// Provider implements the Anthropic provider
type Provider struct {
    config     *Config
    httpClient *http.Client
    logger     aiclient.Logger
}

// Config represents Anthropic-specific configuration
type Config struct {
    APIKey       string
    BaseURL      string // defaults to https://api.anthropic.com
    DefaultModel string
    Version      string // API version
}

// NewProvider creates a new Anthropic provider
func NewProvider(cfg *Config, opts ...Option) (*Provider, error) {
    if cfg.BaseURL == "" {
        cfg.BaseURL = "https://api.anthropic.com"
    }

    if cfg.DefaultModel == "" {
        cfg.DefaultModel = "claude-3-5-sonnet-20241022"
    }

    if cfg.Version == "" {
        cfg.Version = "2023-06-01"
    }

    return &Provider{
        config:     cfg,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }, nil
}

// Chat converts universal format to Anthropic's Messages API
func (p *Provider) Chat(ctx context.Context, req *aiclient.ChatRequest) (*aiclient.ChatResponse, error) {
    // Handle system message separately (Anthropic uses different format)
    // Convert to Anthropic Messages API format
}
```

## Middleware System

### Built-in Middleware

```go
// LoggingMiddleware logs all requests and responses
func LoggingMiddleware(logger Logger) Middleware {
    return func(next Provider) Provider {
        return &loggingProvider{
            next:   next,
            logger: logger,
        }
    }
}

type loggingProvider struct {
    next   Provider
    logger Logger
}

func (p *loggingProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    start := time.Now()

    p.logger.Info("AI request started",
        "provider", p.next.Name(),
        "messages", len(req.Messages),
    )

    resp, err := p.next.Chat(ctx, req)

    duration := time.Since(start)

    if err != nil {
        p.logger.Error("AI request failed",
            "provider", p.next.Name(),
            "duration", duration,
            "error", err,
        )
    } else {
        p.logger.Info("AI request completed",
            "provider", p.next.Name(),
            "duration", duration,
            "tokens", resp.Usage.TotalTokens,
        )
    }

    return resp, err
}

// RetryMiddleware adds retry logic with exponential backoff
func RetryMiddleware(maxRetries int, backoff time.Duration) Middleware {
    return func(next Provider) Provider {
        return &retryProvider{
            next:       next,
            maxRetries: maxRetries,
            backoff:    backoff,
        }
    }
}

// CircuitBreakerMiddleware adds circuit breaker pattern
func CircuitBreakerMiddleware(threshold int, timeout time.Duration) Middleware {
    return func(next Provider) Provider {
        return &circuitBreakerProvider{
            next:      next,
            threshold: threshold,
            timeout:   timeout,
        }
    }
}

// RateLimitMiddleware adds rate limiting
func RateLimitMiddleware(requestsPerSecond float64) Middleware {
    return func(next Provider) Provider {
        return &rateLimitProvider{
            next:    next,
            limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
        }
    }
}

// CachingMiddleware adds response caching
func CachingMiddleware(cache Cache) Middleware {
    return func(next Provider) Provider {
        return &cachingProvider{
            next:  next,
            cache: cache,
        }
    }
}

// MetricsMiddleware adds metrics collection
func MetricsMiddleware(metrics MetricsCollector) Middleware {
    return func(next Provider) Provider {
        return &metricsProvider{
            next:    next,
            metrics: metrics,
        }
    }
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yourusername/core-backend/pkg/aiclient"
    "github.com/yourusername/core-backend/pkg/aiclient/providers/openai"
)

func main() {
    // Load configuration from environment
    cfg, err := aiclient.LoadFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    // Create client
    client, err := aiclient.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Use fluent API
    resp, err := client.NewRequest().
        System("You are a helpful assistant.").
        User("What is the capital of France?").
        Model("gpt-4").
        Temperature(0.7).
        Do(context.Background())

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Message.Content)
}
```

### Streaming Example

```go
func streamingExample() {
    client, _ := aiclient.NewClient(cfg)
    defer client.Close()

    stream, err := client.NewRequest().
        User("Write a short story about a robot.").
        Stream(context.Background())

    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()

    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

        fmt.Print(chunk.Delta.Content)
    }
}
```

### Multi-Provider Example

```go
func multiProviderExample() {
    // OpenAI client
    openaiClient, _ := aiclient.NewClient(&aiclient.Config{
        Provider: "openai",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    // Anthropic client
    anthropicClient, _ := aiclient.NewClient(&aiclient.Config{
        Provider: "anthropic",
        APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
    })

    // Use both
    resp1, _ := openaiClient.NewRequest().User("Hello").Do(ctx)
    resp2, _ := anthropicClient.NewRequest().User("Hello").Do(ctx)

    fmt.Println("OpenAI:", resp1.Message.Content)
    fmt.Println("Anthropic:", resp2.Message.Content)
}
```

### Middleware Example

```go
func middlewareExample() {
    cfg, _ := aiclient.LoadFromEnv()
    client, _ := aiclient.NewClient(cfg)

    // Add middleware
    client.Use(
        aiclient.LoggingMiddleware(logger),
        aiclient.RetryMiddleware(3, time.Second),
        aiclient.CircuitBreakerMiddleware(5, 30*time.Second),
        aiclient.RateLimitMiddleware(10.0), // 10 req/s
    )

    resp, _ := client.NewRequest().User("Hello").Do(ctx)
}
```

### Integration with Server Package

```go
package main

import (
    "context"

    "github.com/yourusername/core-backend/pkg/aiclient"
    "github.com/yourusername/core-backend/pkg/server"

    pb "your/proto/package"
)

// AIService implements gRPC service with AI client
type AIService struct {
    pb.UnimplementedAIServiceServer
    aiClient aiclient.Client
}

func (s *AIService) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
    // Use AI client
    resp, err := s.aiClient.NewRequest().
        System(req.SystemPrompt).
        User(req.Message).
        Model(req.Model).
        Do(ctx)

    if err != nil {
        return nil, err
    }

    return &pb.ChatResponse{
        Message: resp.Message.Content,
        Tokens:  int32(resp.Usage.TotalTokens),
    }, nil
}

func main() {
    // Create AI client
    aiCfg, _ := aiclient.LoadFromEnv()
    aiClient, _ := aiclient.NewClient(aiCfg)

    // Create server
    srv := server.NewServer(
        server.WithAddress(":8080"),
        server.WithGRPCAddress(":9090"),
    )

    // Register AI service
    aiService := &AIService{aiClient: aiClient}
    pb.RegisterAIServiceServer(srv.GRPCServer(), aiService)

    // Start server
    srv.Start()
}
```

## Testing Strategy

### Mock Provider

```go
package testutil

// MockProvider is a mock implementation for testing
type MockProvider struct {
    ChatFunc       func(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStreamFunc func(ctx context.Context, req *ChatRequest) (ChatStream, error)
    NameFunc       func() string
}

func (m *MockProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    if m.ChatFunc != nil {
        return m.ChatFunc(ctx, req)
    }
    return &ChatResponse{
        Message: Message{
            Role:    RoleAssistant,
            Content: "Mock response",
        },
    }, nil
}
```

### Test Example

```go
func TestClient_Chat(t *testing.T) {
    mockProvider := &testutil.MockProvider{
        ChatFunc: func(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
            return &ChatResponse{
                Message: Message{
                    Role:    RoleAssistant,
                    Content: "Test response",
                },
            }, nil
        },
    }

    client := NewClientWithProvider(mockProvider)

    resp, err := client.NewRequest().
        User("Test message").
        Do(context.Background())

    assert.NoError(t, err)
    assert.Equal(t, "Test response", resp.Message.Content)
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- [ ] Define core interfaces (`Client`, `Provider`, `ChatStream`)
- [ ] Implement universal domain models (`Message`, `ChatRequest`, `ChatResponse`)
- [ ] Create configuration system with env loading
- [ ] Implement request builder (fluent API)
- [ ] Set up error definitions and handling
- [ ] Create logger interface
- [ ] Write comprehensive tests for core components

### Phase 2: Provider Implementations (Week 2)
- [ ] Implement OpenAI provider
  - [ ] Chat completions
  - [ ] Streaming support
  - [ ] Embeddings
  - [ ] Error handling and rate limits
- [ ] Implement Anthropic provider
  - [ ] Messages API integration
  - [ ] Streaming support
  - [ ] Handle system messages correctly
- [ ] Implement Google AI provider (Gemini)
- [ ] Implement Azure OpenAI provider

### Phase 3: Resilience Features (Week 3)
- [ ] Implement retry logic with exponential backoff
- [ ] Implement circuit breaker pattern
- [ ] Implement rate limiting
- [ ] Add timeout handling
- [ ] Create middleware system
- [ ] Implement built-in middleware (logging, retry, CB, rate limit)

### Phase 4: Advanced Features (Week 4)
- [ ] Implement streaming support across all providers
- [ ] Add tool/function calling support
- [ ] Implement response caching middleware
- [ ] Add metrics collection middleware
- [ ] Create provider fallback mechanism
- [ ] Add context propagation (tracing, cancellation)

### Phase 5: Testing & Documentation (Week 5)
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests for each provider
- [ ] Create mock implementations for testing
- [ ] Write comprehensive README.md
- [ ] Create usage examples (basic, streaming, multi-provider, middleware)
- [ ] Add godoc comments to all public APIs
- [ ] Create migration guide for switching providers

### Phase 6: Integration & Polish (Week 6)
- [ ] Create example integrating with `server` package
- [ ] Create example integrating with `auth` package (API key validation)
- [ ] Add benchmarks
- [ ] Performance optimization
- [ ] Security audit
- [ ] Final documentation review

## Error Handling

### Error Types

```go
var (
    // Configuration errors
    ErrInvalidProvider = errors.New("invalid provider")
    ErrMissingAPIKey   = errors.New("missing API key")
    ErrInvalidConfig   = errors.New("invalid configuration")

    // Request errors
    ErrEmptyMessages    = errors.New("empty messages")
    ErrInvalidMessage   = errors.New("invalid message format")
    ErrInvalidModel     = errors.New("invalid model")

    // Response errors
    ErrRateLimitExceeded     = errors.New("rate limit exceeded")
    ErrQuotaExceeded         = errors.New("quota exceeded")
    ErrInvalidRequest        = errors.New("invalid request")
    ErrAuthenticationFailed  = errors.New("authentication failed")
    ErrModelNotFound         = errors.New("model not found")
    ErrServerError           = errors.New("server error")

    // Streaming errors
    ErrStreamClosed = errors.New("stream closed")
    ErrStreamError  = errors.New("stream error")

    // Circuit breaker errors
    ErrCircuitOpen = errors.New("circuit breaker open")
)

// ProviderError wraps provider-specific errors
type ProviderError struct {
    Provider   string
    StatusCode int
    Message    string
    Err        error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("%s provider error [%d]: %s", e.Provider, e.StatusCode, e.Message)
}

func (e *ProviderError) Unwrap() error {
    return e.Err
}
```

## Security Considerations

### API Key Management

```go
// Never log API keys
func sanitizeConfig(cfg *Config) *Config {
    sanitized := *cfg
    if sanitized.APIKey != "" {
        sanitized.APIKey = "***"
    }
    return &sanitized
}

// Use environment variables, not hardcoded keys
// Support secret managers (AWS Secrets Manager, etc.)
```

### Input Validation

```go
// Validate all inputs
func validateChatRequest(req *ChatRequest) error {
    if len(req.Messages) == 0 {
        return ErrEmptyMessages
    }

    for i, msg := range req.Messages {
        if msg.Role == "" {
            return fmt.Errorf("message %d: %w", i, ErrInvalidMessage)
        }
        if msg.Content == "" && len(msg.ToolCalls) == 0 {
            return fmt.Errorf("message %d: %w", i, ErrInvalidMessage)
        }
    }

    return nil
}
```

### Rate Limiting

```go
// Prevent abuse with rate limiting
client.Use(aiclient.RateLimitMiddleware(10.0)) // 10 req/s
```

## Performance Optimization

### Connection Pooling

```go
// Reuse HTTP connections
func newHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}
```

### Context Propagation

```go
// Always respect context cancellation
func (p *Provider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    // Create HTTP request with context
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
    if err != nil {
        return nil, err
    }

    // Context cancellation will cancel the request
    resp, err := p.httpClient.Do(httpReq)
    // ...
}
```

### Caching

```go
// Cache responses for identical requests
type Cache interface {
    Get(key string) (*ChatResponse, bool)
    Set(key string, resp *ChatResponse, ttl time.Duration)
}

func CachingMiddleware(cache Cache) Middleware {
    return func(next Provider) Provider {
        return &cachingProvider{next: next, cache: cache}
    }
}
```

## Metrics and Observability

### Metrics Interface

```go
// MetricsCollector defines metrics collection interface
type MetricsCollector interface {
    // RecordRequest records a request
    RecordRequest(provider, model string)

    // RecordResponse records a response
    RecordResponse(provider, model string, duration time.Duration, tokens int, err error)

    // RecordStreamChunk records a stream chunk
    RecordStreamChunk(provider, model string)
}
```

### Logging Best Practices

```go
// Use structured logging
logger.Info("AI request",
    "provider", provider,
    "model", model,
    "messages", len(messages),
    "tokens", tokens,
    "duration_ms", duration.Milliseconds(),
)

// Never log sensitive data (API keys, user content unless explicitly enabled)
```

## Documentation Requirements

### README.md Structure

1. **Overview** - What is this package?
2. **Installation** - How to install
3. **Quick Start** - Simple example
4. **Supported Providers** - List with links
5. **Configuration** - Environment variables
6. **Usage Examples** - Multiple scenarios
7. **Middleware** - Available middleware
8. **Error Handling** - Error types and handling
9. **Testing** - How to test with mocks
10. **Best Practices** - Security, performance
11. **API Reference** - Link to godoc
12. **Contributing** - How to add providers

### GoDoc Comments

```go
// Package aiclient provides a provider-agnostic interface for interacting
// with AI providers such as OpenAI, Anthropic, Google AI, and Azure OpenAI.
//
// # Basic Usage
//
//     cfg, _ := aiclient.LoadFromEnv()
//     client, _ := aiclient.NewClient(cfg)
//
//     resp, _ := client.NewRequest().
//         User("Hello, world!").
//         Do(context.Background())
//
//     fmt.Println(resp.Message.Content)
//
// # Streaming
//
//     stream, _ := client.NewRequest().
//         User("Write a story").
//         Stream(context.Background())
//
//     for {
//         chunk, err := stream.Recv()
//         if err == io.EOF {
//             break
//         }
//         fmt.Print(chunk.Delta.Content)
//     }
//
// For more examples, see the examples/ directory.
package aiclient
```

## Migration and Compatibility

### Provider Migration

```go
// Easy to switch providers via environment variable
// Before: AI_PROVIDER=openai
// After:  AI_PROVIDER=anthropic

// Code doesn't change!
client, _ := aiclient.LoadFromEnv()
resp, _ := client.NewRequest().User("Hello").Do(ctx)
```

### Version Compatibility

```go
// Support provider API versioning
type ProviderConfig interface {
    Version() string
    SetVersion(version string)
}

// OpenAI: v1
// Anthropic: 2023-06-01
```

## Success Criteria

1. ✅ **Provider Agnostic**: Works with 4+ providers (OpenAI, Anthropic, Google, Azure)
2. ✅ **Easy to Use**: Fluent API, minimal configuration
3. ✅ **Production Ready**: Retry, circuit breaker, rate limiting
4. ✅ **Well Tested**: >80% test coverage
5. ✅ **Well Documented**: Comprehensive README, examples, godoc
6. ✅ **Performant**: Connection pooling, caching support
7. ✅ **Observable**: Logging, metrics interfaces
8. ✅ **Secure**: API key management, input validation
9. ✅ **Extensible**: Middleware system for custom logic
10. ✅ **Consistent**: Follows existing codebase patterns

## Open Questions

1. **Tool/Function Calling**: Each provider has different formats - how to unify?
2. **Embeddings**: Some providers have different endpoints - create separate interface?
3. **Image Support**: Vision models - extend Message struct or separate interface?
4. **Fine-tuning**: Provider-specific - out of scope for v1?
5. **Prompt Caching**: Anthropic supports prompt caching - how to expose in universal API?
6. **Batch API**: OpenAI batch API - separate interface or part of main client?

## Future Enhancements

1. **Provider Auto-Failover**: Automatically switch to backup provider on failure
2. **Cost Tracking**: Track costs per request based on token usage
3. **Prompt Templates**: Built-in template system for common patterns
4. **Response Validation**: JSON schema validation for structured outputs
5. **A/B Testing**: Route requests to different providers for comparison
6. **Local Models**: Support for local LLMs (Ollama, llama.cpp)
7. **Vector Embeddings**: First-class support for embeddings and vector search
8. **Multi-modal**: Image, audio, video support

## Conclusion

This plan provides a comprehensive roadmap for building a production-ready, provider-agnostic AI client library that follows the architectural patterns established in the core-backend codebase. The design prioritizes:

- **Simplicity**: Easy to use, minimal configuration
- **Flexibility**: Support multiple providers, extensible via middleware
- **Reliability**: Retry logic, circuit breaker, rate limiting
- **Testability**: Mock implementations, comprehensive tests
- **Observability**: Logging, metrics, tracing support

The implementation will be done in phases over 6 weeks, with each phase building on the previous one. The result will be a robust, well-documented package that can be used across multiple projects.
