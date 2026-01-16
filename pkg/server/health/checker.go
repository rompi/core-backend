package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// --- Built-in Health Checkers ---

// PingChecker returns a checker that always succeeds.
// Useful as a baseline health check.
func PingChecker() CheckerFunc {
	return func(ctx context.Context) error {
		return nil
	}
}

// DatabaseChecker returns a checker for database connectivity.
type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

func DatabaseChecker(db DatabasePinger) CheckerFunc {
	return func(ctx context.Context) error {
		return db.PingContext(ctx)
	}
}

// SQLDBChecker returns a checker for *sql.DB.
func SQLDBChecker(db *sql.DB) CheckerFunc {
	return func(ctx context.Context) error {
		return db.PingContext(ctx)
	}
}

// GRPCChecker returns a checker for gRPC connection state.
func GRPCChecker(conn *grpc.ClientConn) CheckerFunc {
	return func(ctx context.Context) error {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.Idle {
			// Trigger connection attempt
			conn.Connect()
			return nil
		}
		return fmt.Errorf("gRPC connection not ready: %s", state.String())
	}
}

// HTTPChecker returns a checker that performs an HTTP GET request.
func HTTPChecker(url string, timeout time.Duration) CheckerFunc {
	client := &http.Client{
		Timeout: timeout,
	}

	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP request returned status %d", resp.StatusCode)
		}

		return nil
	}
}

// TCPChecker returns a checker that attempts a TCP connection.
func TCPChecker(addr string, timeout time.Duration) CheckerFunc {
	return func(ctx context.Context) error {
		deadline := time.Now().Add(timeout)
		ctxWithTimeout, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		var d net.Dialer
		conn, err := d.DialContext(ctxWithTimeout, "tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP connection failed: %w", err)
		}
		conn.Close()
		return nil
	}
}

// RedisPinger is an interface for Redis ping operations.
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// RedisChecker returns a checker for Redis connectivity.
func RedisChecker(client RedisPinger) CheckerFunc {
	return func(ctx context.Context) error {
		return client.Ping(ctx)
	}
}

// TimeoutChecker wraps a checker with a timeout.
func TimeoutChecker(check CheckerFunc, timeout time.Duration) CheckerFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- check(ctx)
		}()

		select {
		case <-ctx.Done():
			return fmt.Errorf("health check timed out after %s", timeout)
		case err := <-done:
			return err
		}
	}
}

// ThresholdChecker fails only after N consecutive failures.
type ThresholdChecker struct {
	check     CheckerFunc
	threshold int
	failures  int
	mu        sync.Mutex
}

// NewThresholdChecker creates a threshold-based checker.
func NewThresholdChecker(check CheckerFunc, threshold int) CheckerFunc {
	tc := &ThresholdChecker{
		check:     check,
		threshold: threshold,
	}
	return tc.Check
}

func (tc *ThresholdChecker) Check(ctx context.Context) error {
	err := tc.check(ctx)

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err != nil {
		tc.failures++
		if tc.failures >= tc.threshold {
			return fmt.Errorf("health check failed %d times: %w", tc.failures, err)
		}
		return nil // Don't fail yet
	}

	tc.failures = 0
	return nil
}

// CompositeChecker combines multiple checkers.
func CompositeChecker(checks ...CheckerFunc) CheckerFunc {
	return func(ctx context.Context) error {
		for _, check := range checks {
			if err := check(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
