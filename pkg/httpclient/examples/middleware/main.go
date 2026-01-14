package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rompi/core-backend/pkg/httpclient"
)

func main() {
	// Create a client with default configuration
	client := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Example 1: Built-in authentication middleware
	fmt.Println("=== Example 1: Bearer Token Authentication ===")
	authExample(client)

	// Example 2: API Key authentication
	fmt.Println("\n=== Example 2: API Key Authentication ===")
	apiKeyExample(client)

	// Example 3: User-Agent middleware
	fmt.Println("\n=== Example 3: Custom User-Agent ===")
	userAgentExample(client)

	// Example 4: Custom headers middleware
	fmt.Println("\n=== Example 4: Custom Headers ===")
	customHeadersExample(client)

	// Example 5: Logging middleware
	fmt.Println("\n=== Example 5: Request/Response Logging ===")
	loggingExample()

	// Example 6: Custom middleware
	fmt.Println("\n=== Example 6: Custom Middleware ===")
	customMiddlewareExample()

	// Example 7: Chaining multiple middleware
	fmt.Println("\n=== Example 7: Middleware Chain ===")
	middlewareChainExample()
}

func authExample(client *httpclient.Client) {
	// Add Bearer token authentication
	client.Use(httpclient.AuthBearerMiddleware("my-secret-token-12345"))

	ctx := context.Background()
	resp, err := client.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	// The Authorization header is automatically added
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Request included: Authorization: Bearer my-secret-token-12345")
}

func apiKeyExample(client *httpclient.Client) {
	// Create a new client for this example to avoid middleware conflicts
	apiClient := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Add API key authentication
	apiClient.Use(httpclient.AuthAPIKeyMiddleware("X-API-Key", "api-key-xyz-789"))

	ctx := context.Background()
	resp, err := apiClient.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Request included: X-API-Key: api-key-xyz-789")
}

func userAgentExample(client *httpclient.Client) {
	// Create a new client for this example
	uaClient := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Add custom user agent
	uaClient.Use(httpclient.UserAgentMiddleware("MyApp/2.0 (Go httpclient)"))

	ctx := context.Background()
	resp, err := uaClient.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Request included: User-Agent: MyApp/2.0 (Go httpclient)")
}

func customHeadersExample(client *httpclient.Client) {
	// Create a new client for this example
	headerClient := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Add multiple custom headers
	headerClient.Use(httpclient.HeaderMiddleware(map[string]string{
		"X-API-Version":  "v2",
		"X-Client-ID":    "client-123",
		"X-Request-Time": time.Now().Format(time.RFC3339),
	}))

	ctx := context.Background()
	resp, err := headerClient.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Request included custom headers:")
	fmt.Println("  X-API-Version: v2")
	fmt.Println("  X-Client-ID: client-123")
	fmt.Println("  X-Request-Time: [timestamp]")
}

func loggingExample() {
	// Create a simple logger
	logger := NewSimpleLogger()

	// Create client with logger
	logClient, err := httpclient.New(httpclient.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Logger:  logger,
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	// Add logging middleware
	logClient.Use(httpclient.LoggingMiddleware(logger))

	ctx := context.Background()
	resp, err := logClient.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d (check logs above for request/response details)\n", resp.StatusCode)
}

func customMiddlewareExample() {
	// Create a client
	client := httpclient.NewDefault("https://jsonplaceholder.typicode.com")

	// Add custom request ID middleware
	client.Use(RequestIDMiddleware())

	// Add custom timing middleware
	client.Use(TimingMiddleware())

	ctx := context.Background()
	resp, err := client.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Custom middleware added request ID and measured timing")
}

func middlewareChainExample() {
	logger := NewSimpleLogger()

	// Create client with configuration
	client, err := httpclient.New(httpclient.Config{
		BaseURL:      "https://jsonplaceholder.typicode.com",
		MaxRetries:   2,
		RetryWaitMin: 500 * time.Millisecond,
		Logger:       logger,
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	// Chain multiple middleware (executed in order)
	client.Use(RequestIDMiddleware())                                   // 1. Add request ID
	client.Use(httpclient.AuthBearerMiddleware("token-abc123"))         // 2. Add auth
	client.Use(httpclient.UserAgentMiddleware("MyApp/1.0"))             // 3. Set user agent
	client.Use(httpclient.HeaderMiddleware(map[string]string{           // 4. Add custom headers
		"X-API-Version": "v1",
	}))
	client.Use(TimingMiddleware())                                      // 5. Measure timing
	client.Use(httpclient.LoggingMiddleware(logger))                    // 6. Log everything

	ctx := context.Background()
	resp, err := client.Post(ctx, "/posts").
		JSON(map[string]string{
			"title": "Middleware Chain Example",
			"body":  "This request went through 6 middleware layers",
		}).
		Do()

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Println("Request processed through complete middleware chain")
}

// Custom middleware: Request ID
// Note: We use a type that implements http.RoundTripper
func RequestIDMiddleware() httpclient.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return &requestIDRoundTripper{next: next}
	}
}

type requestIDRoundTripper struct {
	next http.RoundTripper
}

func (rt *requestIDRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Generate a simple request ID (in production, use UUID or similar)
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Add to request header
	req.Header.Set("X-Request-ID", requestID)

	fmt.Printf("[RequestID] Added ID: %s\n", requestID)

	// Execute the request
	return rt.next.RoundTrip(req)
}

// Custom middleware: Timing
// Note: We use a type that implements http.RoundTripper
func TimingMiddleware() httpclient.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return &timingRoundTripper{next: next}
	}
}

type timingRoundTripper struct {
	next http.RoundTripper
}

func (rt *timingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Start timing
	start := time.Now()

	// Execute the request
	resp, err := rt.next.RoundTrip(req)

	// Calculate duration
	duration := time.Since(start)

	fmt.Printf("[Timing] Request to %s took %v\n", req.URL.Path, duration)

	// Add timing header to response
	if resp != nil {
		resp.Header.Set("X-Response-Time", duration.String())
	}

	return resp, err
}

// SimpleLogger implements the httpclient.Logger interface
type SimpleLogger struct{}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{}
}

func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log("DEBUG", msg, keysAndValues...)
}

func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log("INFO", msg, keysAndValues...)
}

func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log("WARN", msg, keysAndValues...)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log("ERROR", msg, keysAndValues...)
}

func (l *SimpleLogger) log(level, msg string, keysAndValues ...interface{}) {
	fmt.Printf("[%s] %s", level, msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
		}
	}
	fmt.Println()
}
