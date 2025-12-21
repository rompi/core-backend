package httpclient

import (
	"errors"
	"net/http"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		want    string
	}{
		{
			name: "error with underlying error",
			err: &Error{
				StatusCode: 500,
				Message:    "server error",
				Err:        errors.New("connection failed"),
			},
			want: "httpclient: server error (status: 500): connection failed",
		},
		{
			name: "error without underlying error",
			err: &Error{
				StatusCode: 404,
				Message:    "not found",
			},
			want: "httpclient: not found (status: 404)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &Error{
		StatusCode: 500,
		Message:    "server error",
		Err:        underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("Error.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *Error
		target error
		want   bool
	}{
		{
			name: "matches sentinel error",
			err: &Error{
				StatusCode: 500,
				Message:    "timeout",
				Err:        ErrTimeout,
			},
			target: ErrTimeout,
			want:   true,
		},
		{
			name: "does not match different sentinel error",
			err: &Error{
				StatusCode: 500,
				Message:    "timeout",
				Err:        ErrTimeout,
			},
			target: ErrCircuitOpen,
			want:   false,
		},
		{
			name: "no underlying error",
			err: &Error{
				StatusCode: 500,
				Message:    "server error",
			},
			target: ErrTimeout,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Is(tt.target)
			if got != tt.want {
				t.Errorf("Error.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"timeout", ErrTimeout, "httpclient: request timeout"},
		{"circuit open", ErrCircuitOpen, "httpclient: circuit breaker open"},
		{"max retries", ErrMaxRetriesExceeded, "httpclient: max retries exceeded"},
		{"invalid config", ErrInvalidConfig, "httpclient: invalid configuration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestError_WithRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	resp := &http.Response{StatusCode: 500}

	err := &Error{
		StatusCode: 500,
		Message:    "server error",
		Body:       []byte("error details"),
		Request:    req,
		Response:   resp,
		Err:        errors.New("connection failed"),
	}

	if err.Request != req {
		t.Error("Request not properly stored")
	}

	if err.Response != resp {
		t.Error("Response not properly stored")
	}

	if string(err.Body) != "error details" {
		t.Errorf("Body = %q, want %q", string(err.Body), "error details")
	}
}
