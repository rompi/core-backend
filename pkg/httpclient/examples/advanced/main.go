package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/rompi/core-backend/pkg/httpclient"
)

func main() {
	// Example 1: Retry policy with exponential backoff
	fmt.Println("=== Example 1: Retry Policy ===")
	retryExample()

	// Example 2: Circuit breaker pattern
	fmt.Println("\n=== Example 2: Circuit Breaker ===")
	circuitBreakerExample()

	// Example 3: Error handling
	fmt.Println("\n=== Example 3: Error Handling ===")
	errorHandlingExample()

	// Example 4: Context cancellation and timeout
	fmt.Println("\n=== Example 4: Context Cancellation ===")
	contextExample()

	// Example 5: Custom retry policy
	fmt.Println("\n=== Example 5: Custom Configuration ===")
	customConfigExample()

	// Example 6: Response status helpers
	fmt.Println("\n=== Example 6: Response Status Helpers ===")
	responseStatusExample()

	// Example 7: Advanced request building
	fmt.Println("\n=== Example 7: Advanced Request Building ===")
	advancedRequestExample()
}

func retryExample() {
	// Create a test server that fails a few times then succeeds
	attemptCount := atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attemptCount.Add(1)
		if count < 3 {
			// Fail the first 2 attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Printf("  [Server] Attempt %d: Returning 503\n", count)
			return
		}
		// Succeed on 3rd attempt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
		fmt.Printf("  [Server] Attempt %d: Returning 200\n", count)
	}))
	defer server.Close()

	// Create client with retry configuration
	client, err := httpclient.New(httpclient.Config{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 1 * time.Second,
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/api/data").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Final status: %d (succeeded after retries)\n", resp.StatusCode)
	fmt.Printf("Total attempts: %d\n", attemptCount.Load())
}

func circuitBreakerExample() {
	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with circuit breaker
	client, err := httpclient.New(httpclient.Config{
		BaseURL:    server.URL,
		MaxRetries: 0, // Disable retries to see circuit breaker in action
		CircuitBreaker: &httpclient.CircuitBreakerConfig{
			MaxRequests: 2,
			Interval:    10 * time.Second,
			Timeout:     5 * time.Second,
			ReadyToTrip: func(counts httpclient.Counts) bool {
				// Open circuit after 3 consecutive failures
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 3 && failureRatio >= 0.6
			},
		},
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Make multiple requests to trigger circuit breaker
	for i := 1; i <= 5; i++ {
		resp, err := client.Get(ctx, "/api/data").Do()

		if err != nil {
			if errors.Is(err, httpclient.ErrCircuitOpen) {
				fmt.Printf("Request %d: Circuit breaker is OPEN (fast fail)\n", i)
			} else {
				fmt.Printf("Request %d: Failed with error: %v\n", i, err)
			}
			continue
		}

		if resp != nil {
			resp.Body.Close()
			fmt.Printf("Request %d: Status %d\n", i, resp.StatusCode)
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
}

func errorHandlingExample() {
	// Create a test server with various error scenarios
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Resource not found"}`))
		case "/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Internal server error"}`))
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Unauthorized"}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client, err := httpclient.New(httpclient.Config{
		BaseURL:    server.URL,
		MaxRetries: 0, // Disable retries for this example
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Test different error scenarios
	testCases := []struct {
		path     string
		desc     string
	}{
		{"/not-found", "404 Not Found"},
		{"/unauthorized", "401 Unauthorized"},
		{"/server-error", "500 Server Error"},
	}

	for _, tc := range testCases {
		resp, err := client.Get(ctx, tc.path).Do()

		fmt.Printf("\nTesting %s:\n", tc.desc)

		if err != nil {
			// Check for specific error types
			var httpErr *httpclient.Error
			if errors.As(err, &httpErr) {
				fmt.Printf("  HTTP Error: %d - %s\n", httpErr.StatusCode, httpErr.Message)
			} else {
				fmt.Printf("  Error: %v\n", err)
			}
			continue
		}

		if resp != nil {
			defer resp.Body.Close()

			// Use response status helpers
			if resp.IsClientError() {
				fmt.Printf("  Client error (4xx): %d\n", resp.StatusCode)
			}
			if resp.IsServerError() {
				fmt.Printf("  Server error (5xx): %d\n", resp.StatusCode)
			}
			if resp.IsSuccess() {
				fmt.Printf("  Success (2xx): %d\n", resp.StatusCode)
			}

			// Get error message from response body
			body, _ := resp.String()
			fmt.Printf("  Response body: %s\n", body)
		}
	}
}

func contextExample() {
	// Create a slow test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.NewDefault(server.URL)

	// Example 1: Context with timeout
	fmt.Println("Test 1: Request with 1 second timeout (should timeout)")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel1()

	start := time.Now()
	_, err := client.Get(ctx1, "/slow").Do()
	duration := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("  Request timed out after %v (as expected)\n", duration)
		} else {
			fmt.Printf("  Error: %v\n", err)
		}
	}

	// Example 2: Manual cancellation
	fmt.Println("\nTest 2: Request with manual cancellation")
	ctx2, cancel2 := context.WithCancel(context.Background())

	// Cancel after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("  Cancelling request...")
		cancel2()
	}()

	start = time.Now()
	_, err = client.Get(ctx2, "/slow").Do()
	duration = time.Since(start)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Printf("  Request cancelled after %v (as expected)\n", duration)
		} else {
			fmt.Printf("  Error: %v\n", err)
		}
	}
}

func customConfigExample() {
	// Create a client with custom configuration
	logger := &SimpleLogger{}

	client, err := httpclient.New(httpclient.Config{
		BaseURL:      "https://jsonplaceholder.typicode.com",
		Timeout:      15 * time.Second,
		MaxRetries:   5,
		RetryWaitMin: 2 * time.Second,
		RetryWaitMax: 30 * time.Second,
		Logger:       logger,
		CircuitBreaker: &httpclient.CircuitBreakerConfig{
			MaxRequests: 3,
			Interval:    30 * time.Second,
			Timeout:     60 * time.Second,
			ReadyToTrip: func(counts httpclient.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 5 && failureRatio >= 0.7
			},
		},
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/posts/1").Do()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully configured client with:\n")
	fmt.Printf("  - 15 second timeout\n")
	fmt.Printf("  - 5 max retries\n")
	fmt.Printf("  - 2-30 second retry backoff\n")
	fmt.Printf("  - Circuit breaker enabled\n")
	fmt.Printf("  - Custom logger\n")
	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

func responseStatusExample() {
	// Create a test server with different status codes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result":"success"}`))
		case "/created":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":123}`))
		case "/no-content":
			w.WriteHeader(http.StatusNoContent)
		case "/redirect":
			w.WriteHeader(http.StatusFound)
			w.Header().Set("Location", "/success")
		case "/bad-request":
			w.WriteHeader(http.StatusBadRequest)
		case "/server-error":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, err := httpclient.New(httpclient.Config{
		BaseURL:         server.URL,
		MaxRetries:      0,
		FollowRedirects: false, // Don't follow redirects for this example
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	testPaths := []string{"/success", "/created", "/no-content", "/redirect", "/bad-request", "/server-error"}

	for _, path := range testPaths {
		resp, err := client.Get(ctx, path).Do()
		if err != nil {
			fmt.Printf("%s: Error - %v\n", path, err)
			continue
		}
		defer resp.Body.Close()

		isRedirect := resp.StatusCode >= 300 && resp.StatusCode < 400

		fmt.Printf("%s (HTTP %d):\n", path, resp.StatusCode)
		fmt.Printf("  IsSuccess: %v\n", resp.IsSuccess())
		fmt.Printf("  IsRedirect: %v\n", isRedirect)
		fmt.Printf("  IsClientError: %v\n", resp.IsClientError())
		fmt.Printf("  IsServerError: %v\n", resp.IsServerError())

		// Decode JSON if present
		if resp.IsSuccess() && resp.StatusCode != http.StatusNoContent {
			var result map[string]interface{}
			if err := resp.JSON(&result); err == nil {
				fmt.Printf("  Body: %v\n", result)
			}
		}
		fmt.Println()
	}
}

func advancedRequestExample() {
	// Create a test server that echoes request details
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON encoding
		w.Write([]byte(fmt.Sprintf(`{"method":"%s","received":"ok"}`, r.Method)))
	}))
	defer server.Close()

	client := httpclient.NewDefault(server.URL)
	ctx := context.Background()

	// Example 1: Complex query parameters
	fmt.Println("Request 1: Complex query parameters")
	resp1, err := client.Get(ctx, "/search").
		Query("q", "golang httpclient").
		Query("filter", "recent").
		Query("limit", "10").
		Query("offset", "0").
		QueryParams(map[string]string{
			"sort":  "date",
			"order": "desc",
		}).
		Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		defer resp1.Body.Close()
		fmt.Printf("  Status: %d\n", resp1.StatusCode)
	}

	// Example 2: Multiple headers
	fmt.Println("\nRequest 2: Multiple custom headers")
	resp2, err := client.Post(ctx, "/api/data").
		Header("Authorization", "Bearer token123").
		Header("X-API-Version", "v2").
		Header("X-Request-ID", "req-12345").
		Headers(map[string]string{
			"X-Client-Type":    "web",
			"X-Client-Version": "1.0.0",
		}).
		JSON(map[string]string{
			"data": "example",
		}).
		Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		defer resp2.Body.Close()
		fmt.Printf("  Status: %d\n", resp2.StatusCode)
	}

	// Example 3: PUT request with JSON
	fmt.Println("\nRequest 3: PUT with complex JSON")
	updateData := map[string]interface{}{
		"name":   "Updated Item",
		"active": true,
		"tags":   []string{"important", "reviewed"},
		"metadata": map[string]interface{}{
			"updated_by": "user123",
			"timestamp":  time.Now().Unix(),
		},
	}
	resp3, err := client.Put(ctx, "/items/456").
		Header("Authorization", "Bearer token123").
		JSON(updateData).
		Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		defer resp3.Body.Close()
		fmt.Printf("  Status: %d\n", resp3.StatusCode)
	}

	// Example 4: PATCH request
	fmt.Println("\nRequest 4: PATCH for partial update")
	resp4, err := client.Patch(ctx, "/users/789").
		JSON(map[string]interface{}{
			"status": "active",
		}).
		Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		defer resp4.Body.Close()
		fmt.Printf("  Status: %d\n", resp4.StatusCode)
	}

	// Example 5: DELETE request
	fmt.Println("\nRequest 5: DELETE resource")
	resp5, err := client.Delete(ctx, "/items/456").
		Header("Authorization", "Bearer token123").
		Do()
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		defer resp5.Body.Close()
		fmt.Printf("  Status: %d\n", resp5.StatusCode)
	}
}

// SimpleLogger implements the httpclient.Logger interface
type SimpleLogger struct{}

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
