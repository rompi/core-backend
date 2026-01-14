package httpclient

import (
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestRetryPolicy_ShouldRetry_NetworkErrors(t *testing.T) {
	rp := &RetryPolicy{
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"EOF error", io.EOF, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"connection refused", syscall.ECONNREFUSED, true},
		{"connection reset", syscall.ECONNRESET, true},
		{"broken pipe", syscall.EPIPE, true},
		{"nil error", nil, false},
		{"generic error", errors.New("generic"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rp.ShouldRetry(nil, tt.err)
			if got != tt.want {
				t.Errorf("ShouldRetry(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryPolicy_ShouldRetry_StatusCodes(t *testing.T) {
	rp := &RetryPolicy{
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
	}

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"401 Unauthorized", http.StatusUnauthorized, false},
		{"404 Not Found", http.StatusNotFound, false},
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"504 Gateway Timeout", http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}
			got := rp.ShouldRetry(resp, nil)
			if got != tt.want {
				t.Errorf("ShouldRetry(status=%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestRetryPolicy_ShouldRetry_NoResponseNoError(t *testing.T) {
	rp := &RetryPolicy{
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
	}

	got := rp.ShouldRetry(nil, nil)
	if got {
		t.Error("ShouldRetry(nil, nil) should return false")
	}
}

func TestRetryPolicy_Backoff(t *testing.T) {
	rp := &RetryPolicy{
		MaxRetries:   5,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
	}

	for attempt := 0; attempt < 10; attempt++ {
		duration := rp.Backoff(attempt)

		if duration < 0 {
			t.Errorf("attempt %d: backoff duration is negative: %v", attempt, duration)
		}

		if duration > rp.RetryWaitMax {
			t.Errorf("attempt %d: backoff duration %v exceeds max %v", attempt, duration, rp.RetryWaitMax)
		}
	}
}

func TestRetryPolicy_IsRetryableError_NetError(t *testing.T) {
	rp := &RetryPolicy{}

	// Create a timeout error
	err := &net.DNSError{
		IsTimeout: true,
	}

	if !rp.isRetryableError(err) {
		t.Error("DNS timeout error should be retryable")
	}
}

func TestRetryPolicy_IsRetryableError_OpError(t *testing.T) {
	rp := &RetryPolicy{}

	// Create an OpError
	err := &net.OpError{
		Op:  "dial",
		Err: errors.New("connection refused"),
	}

	if !rp.isRetryableError(err) {
		t.Error("OpError should be retryable")
	}
}
