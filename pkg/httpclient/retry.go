package httpclient

import (
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/rompi/core-backend/pkg/httpclient/internal/backoff"
)

// RetryPolicy defines the retry behavior for failed HTTP requests.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries.
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries.
	RetryWaitMax time.Duration
}

// ShouldRetry determines whether a request should be retried based on
// the response and error. It returns true for transient failures that
// are likely to succeed on retry.
//
// Retryable conditions:
//   - Network errors (connection refused, timeout, etc.)
//   - 5xx server errors
//   - 429 Too Many Requests
//   - Specific transient errors (EOF, broken pipe, etc.)
func (rp *RetryPolicy) ShouldRetry(resp *http.Response, err error) bool {
	// If there's an error, check if it's retryable
	if err != nil {
		return rp.isRetryableError(err)
	}

	// If we have a response, check the status code
	if resp != nil {
		return rp.isRetryableStatusCode(resp.StatusCode)
	}

	return false
}

// Backoff calculates the wait duration before the next retry attempt
// using exponential backoff with jitter.
func (rp *RetryPolicy) Backoff(attempt int) time.Duration {
	return backoff.Exponential(attempt, rp.RetryWaitMin, rp.RetryWaitMax)
}

// isRetryableError checks if an error is retryable.
func (rp *RetryPolicy) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true // includes timeout and temporary errors
	}

	// Connection errors are retryable
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// DNS errors are retryable
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// Specific retryable errors
	switch {
	case errors.Is(err, io.EOF):
		return true
	case errors.Is(err, io.ErrUnexpectedEOF):
		return true
	case errors.Is(err, syscall.ECONNREFUSED):
		return true
	case errors.Is(err, syscall.ECONNRESET):
		return true
	case errors.Is(err, syscall.EPIPE):
		return true
	}

	return false
}

// isRetryableStatusCode checks if an HTTP status code is retryable.
func (rp *RetryPolicy) isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusInternalServerError: // 500
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}
