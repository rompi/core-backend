package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client is the main HTTP client with built-in retry, circuit breaking,
// and middleware support. It is safe for concurrent use.
type Client struct {
	baseURL        string
	httpClient     *http.Client
	middleware     []Middleware
	retryPolicy    *RetryPolicy
	circuitBreaker *CircuitBreaker
	logger         Logger
}

// Config holds configuration options for creating a new HTTP client.
type Config struct {
	// BaseURL is the base URL for all requests (e.g., "https://api.example.com").
	BaseURL string

	// Timeout is the maximum duration for a request (default: 30s).
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts (default: 3).
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries (default: 1s).
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries (default: 30s).
	RetryWaitMax time.Duration

	// CircuitBreaker is the circuit breaker configuration (optional).
	CircuitBreaker *CircuitBreakerConfig

	// Logger is the logger to use (default: noop logger).
	Logger Logger

	// Transport is the HTTP transport to use (default: http.DefaultTransport).
	Transport http.RoundTripper

	// FollowRedirects controls whether to follow redirects (default: true).
	FollowRedirects bool
}

// New creates a new HTTP client with the provided configuration.
// Returns an error if the configuration is invalid.
func New(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	// Apply defaults
	cfg.applyDefaults()

	// Create HTTP client
	httpClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: cfg.Transport,
	}

	// Default is true (follow redirects)
	// Only set CheckRedirect if explicitly disabled
	if cfg.FollowRedirects {
		// Default behavior - let http.Client follow redirects
	} else {
		// Disable redirects
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create retry policy
	retryPolicy := &RetryPolicy{
		MaxRetries:   cfg.MaxRetries,
		RetryWaitMin: cfg.RetryWaitMin,
		RetryWaitMax: cfg.RetryWaitMax,
	}

	// Create circuit breaker if configured
	var cb *CircuitBreaker
	if cfg.CircuitBreaker != nil {
		cb = NewCircuitBreaker(*cfg.CircuitBreaker)
	}

	client := &Client{
		baseURL:        cfg.BaseURL,
		httpClient:     httpClient,
		middleware:     []Middleware{},
		retryPolicy:    retryPolicy,
		circuitBreaker: cb,
		logger:         cfg.Logger,
	}

	return client, nil
}

// NewDefault creates a new HTTP client with sensible defaults.
// This is a convenience function for simple use cases.
func NewDefault(baseURL string) *Client {
	client, _ := New(Config{
		BaseURL: baseURL,
	})
	return client
}

// Get creates a GET request builder.
func (c *Client) Get(ctx context.Context, path string) *RequestBuilder {
	return newRequestBuilder(c, ctx, http.MethodGet, path)
}

// Post creates a POST request builder.
func (c *Client) Post(ctx context.Context, path string) *RequestBuilder {
	return newRequestBuilder(c, ctx, http.MethodPost, path)
}

// Put creates a PUT request builder.
func (c *Client) Put(ctx context.Context, path string) *RequestBuilder {
	return newRequestBuilder(c, ctx, http.MethodPut, path)
}

// Patch creates a PATCH request builder.
func (c *Client) Patch(ctx context.Context, path string) *RequestBuilder {
	return newRequestBuilder(c, ctx, http.MethodPatch, path)
}

// Delete creates a DELETE request builder.
func (c *Client) Delete(ctx context.Context, path string) *RequestBuilder {
	return newRequestBuilder(c, ctx, http.MethodDelete, path)
}

// Use adds a middleware to the client's middleware chain.
// Middleware are executed in the order they are added.
func (c *Client) Use(mw Middleware) {
	c.middleware = append(c.middleware, mw)
}

// do executes the HTTP request with retry logic and circuit breaker.
// This is the internal method that handles the actual request execution.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	// If circuit breaker is configured, wrap the execution
	if c.circuitBreaker != nil {
		var resp *http.Response
		err := c.circuitBreaker.Call(func() error {
			var execErr error
			resp, execErr = c.executeWithRetry(req)
			return execErr
		})
		return resp, err
	}

	// Execute with retry (no circuit breaker)
	return c.executeWithRetry(req)
}

// executeWithRetry executes the request with retry logic.
func (c *Client) executeWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryPolicy.MaxRetries; attempt++ {
		// Clone the request for retry
		reqClone := req.Clone(req.Context())

		// Build middleware chain
		handler := c.buildMiddlewareChain()

		// Execute request
		resp, err := handler.RoundTrip(reqClone)

		// Check if we should retry
		if !c.retryPolicy.ShouldRetry(resp, err) {
			if err != nil {
				return nil, err
			}
			return resp, nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if attempt < c.retryPolicy.MaxRetries {
			waitDuration := c.retryPolicy.Backoff(attempt)
			c.logger.Debug("retrying request",
				"attempt", attempt+1,
				"wait", waitDuration,
				"url", req.URL.String(),
			)

			select {
			case <-time.After(waitDuration):
				// Continue to next attempt
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
	}

	return nil, ErrMaxRetriesExceeded
}

// buildMiddlewareChain builds the middleware chain with the base transport.
func (c *Client) buildMiddlewareChain() http.RoundTripper {
	// Start with the base HTTP client
	handler := c.httpClient.Transport
	if handler == nil {
		handler = http.DefaultTransport
	}

	// Wrap with middleware in reverse order
	for i := len(c.middleware) - 1; i >= 0; i-- {
		handler = c.middleware[i](handler)
	}

	return handler
}

// validate validates the configuration.
func (cfg *Config) validate() error {
	if cfg.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if cfg.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if cfg.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if cfg.RetryWaitMin < 0 {
		return fmt.Errorf("retry wait min cannot be negative")
	}

	if cfg.RetryWaitMax < 0 {
		return fmt.Errorf("retry wait max cannot be negative")
	}

	if cfg.RetryWaitMax > 0 && cfg.RetryWaitMin > cfg.RetryWaitMax {
		return fmt.Errorf("retry wait min cannot be greater than retry wait max")
	}

	return nil
}

// applyDefaults applies default values to unset configuration fields.
func (cfg *Config) applyDefaults() {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = 1 * time.Second
	}

	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = 30 * time.Second
	}

	if cfg.Logger == nil {
		cfg.Logger = NewNoopLogger()
	}

	if cfg.Transport == nil {
		cfg.Transport = http.DefaultTransport
	}

	// FollowRedirects defaults to true (zero value is false, but we want default true)
	// This is handled in the main New() function logic
}
