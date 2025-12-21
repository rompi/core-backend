package httpclient

import (
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker_Defaults(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	if cb.config.MaxRequests != 1 {
		t.Errorf("MaxRequests = %d, want 1", cb.config.MaxRequests)
	}

	if cb.config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want %v", cb.config.Timeout, 60*time.Second)
	}

	if cb.config.ReadyToTrip == nil {
		t.Error("ReadyToTrip function should not be nil")
	}

	if cb.state != StateClosed {
		t.Errorf("initial state = %v, want %v", cb.state, StateClosed)
	}
}

func TestNewCircuitBreaker_CustomConfig(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxRequests: 5,
		Timeout:     30 * time.Second,
		Interval:    10 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	cb := NewCircuitBreaker(config)

	if cb.config.MaxRequests != 5 {
		t.Errorf("MaxRequests = %d, want 5", cb.config.MaxRequests)
	}

	if cb.config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cb.config.Timeout, 30*time.Second)
	}
}

func TestCircuitBreaker_State(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	if cb.State() != StateClosed {
		t.Errorf("initial State() = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreaker_Counts(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	counts := cb.Counts()

	if counts.Requests != 0 {
		t.Errorf("initial Requests = %d, want 0", counts.Requests)
	}

	if counts.TotalSuccesses != 0 {
		t.Errorf("initial TotalSuccesses = %d, want 0", counts.TotalSuccesses)
	}
}

func TestCircuitBreaker_Call_Success(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	counts := cb.Counts()
	if counts.TotalSuccesses != 1 {
		t.Errorf("TotalSuccesses = %d, want 1", counts.TotalSuccesses)
	}

	if counts.ConsecutiveSuccesses != 1 {
		t.Errorf("ConsecutiveSuccesses = %d, want 1", counts.ConsecutiveSuccesses)
	}
}

func TestCircuitBreaker_Call_Failure(t *testing.T) {
	failureCount := 0
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2 // Lower threshold for testing
		},
	})

	testErr := errors.New("test error")

	// First failure - circuit should stay closed
	err := cb.Call(func() error {
		failureCount++
		return testErr
	})

	if err != testErr {
		t.Errorf("call 1: expected test error, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("call 1: state = %v, want %v", cb.State(), StateClosed)
	}

	// Second failure - circuit should open
	err = cb.Call(func() error {
		failureCount++
		return testErr
	})

	if err != testErr {
		t.Errorf("call 2: expected test error, got %v", err)
	}

	if cb.State() != StateOpen {
		t.Errorf("after 2 failures: state = %v, want %v (counts: %+v)", cb.State(), StateOpen, cb.Counts())
	}

	// Next call should be rejected immediately
	err = cb.Call(func() error {
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	// Trigger some successes
	cb.Call(func() error { return nil })
	cb.Call(func() error { return nil })

	cb.Reset()

	counts := cb.Counts()
	if counts.Requests != 0 {
		t.Errorf("Requests after reset = %d, want 0", counts.Requests)
	}

	if counts.TotalSuccesses != 0 {
		t.Errorf("TotalSuccesses after reset = %d, want 0", counts.TotalSuccesses)
	}

	if cb.State() != StateClosed {
		t.Errorf("state after reset = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreaker_String(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	str := cb.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("State.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxRequests: 1,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
	})

	// Trigger failure to open circuit
	testErr := errors.New("test error")
	cb.Call(func() error { return testErr })

	if cb.State() != StateOpen {
		t.Errorf("state = %v, want %v", cb.State(), StateOpen)
	}

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Next call should be allowed (half-open)
	err := cb.Call(func() error { return nil })

	if err != nil {
		t.Errorf("expected success in half-open, got %v", err)
	}

	// Should transition back to closed on success
	if cb.State() != StateClosed {
		t.Errorf("state after success = %v, want %v", cb.State(), StateClosed)
	}
}

func TestCircuitBreaker_MaxRequestsInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		MaxRequests: 2,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
	})

	// Open the circuit
	testErr := errors.New("test error")
	cb.Call(func() error { return testErr })

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// First two requests should be allowed
	cb.Call(func() error { return nil }) // Doesn't matter, just consume request quota

	// Reset to test max requests
	cb.Reset()
	cb.Call(func() error { return testErr }) // Open again
	time.Sleep(150 * time.Millisecond)

	// Manually update state to half-open and max out requests
	cb.mu.Lock()
	cb.state = StateHalfOpen
	cb.counts.Requests = cb.config.MaxRequests
	cb.mu.Unlock()

	// Next request should be rejected
	err := cb.Call(func() error { return nil })
	if err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen when max requests reached, got %v", err)
	}
}
