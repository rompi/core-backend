package httpclient

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for common failure scenarios.
var (
	// ErrTimeout is returned when a request times out.
	ErrTimeout = errors.New("httpclient: request timeout")

	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("httpclient: circuit breaker open")

	// ErrMaxRetriesExceeded is returned when max retry attempts are exhausted.
	ErrMaxRetriesExceeded = errors.New("httpclient: max retries exceeded")

	// ErrInvalidConfig is returned when client configuration is invalid.
	ErrInvalidConfig = errors.New("httpclient: invalid configuration")
)

// Error represents an HTTP error with additional context.
// It wraps the underlying error and provides HTTP-specific information
// such as status code, response body, and request/response objects.
type Error struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Message is a human-readable error message.
	Message string

	// Body contains the raw response body bytes.
	Body []byte

	// Request is the original HTTP request that failed.
	Request *http.Request

	// Response is the HTTP response received.
	Response *http.Response

	// Err is the underlying error that caused this error.
	Err error
}

// Error implements the error interface.
// It returns a formatted error message including the status code and message.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("httpclient: %s (status: %d): %v", e.Message, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("httpclient: %s (status: %d)", e.Message, e.StatusCode)
}

// Unwrap returns the underlying error.
// This allows errors.Is and errors.As to work correctly.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is implements error matching for sentinel errors.
// This enables errors.Is to check if this error wraps a specific sentinel error.
func (e *Error) Is(target error) bool {
	if e.Err == nil {
		return false
	}
	return errors.Is(e.Err, target)
}
