package httpclient

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed means the circuit breaker is closed and requests are allowed.
	StateClosed State = iota

	// StateOpen means the circuit breaker is open and requests are rejected.
	StateOpen

	// StateHalfOpen means the circuit breaker is testing if the service recovered.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed in half-open state.
	// Default: 1
	MaxRequests uint32

	// Interval is the window duration for counting failures.
	// Default: 0 (disabled)
	Interval time.Duration

	// Timeout is the duration the circuit stays open before entering half-open.
	// Default: 60s
	Timeout time.Duration

	// ReadyToTrip is the threshold function to determine if the breaker should trip.
	// It receives the counts in the current window and returns true if the breaker should open.
	// Default: trips after 5 consecutive failures
	ReadyToTrip func(counts Counts) bool
}

// Counts holds the statistics for the circuit breaker.
type Counts struct {
	// Requests is the total number of requests.
	Requests uint32

	// TotalSuccesses is the total number of successful requests.
	TotalSuccesses uint32

	// TotalFailures is the total number of failed requests.
	TotalFailures uint32

	// ConsecutiveSuccesses is the number of consecutive successful requests.
	ConsecutiveSuccesses uint32

	// ConsecutiveFailures is the number of consecutive failed requests.
	ConsecutiveFailures uint32
}

// CircuitBreaker implements the circuit breaker pattern to prevent
// cascading failures. It wraps requests and tracks their success/failure,
// opening the circuit when a failure threshold is reached.
//
// State transitions:
//   - Closed → Open: When ReadyToTrip returns true
//   - Open → Half-Open: After Timeout duration
//   - Half-Open → Closed: When a request succeeds
//   - Half-Open → Open: When a request fails
type CircuitBreaker struct {
	config CircuitBreakerConfig
	state  State
	counts Counts
	expiry time.Time
	mu     sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}

	// Apply defaults
	if cb.config.MaxRequests == 0 {
		cb.config.MaxRequests = 1
	}

	if cb.config.Timeout == 0 {
		cb.config.Timeout = 60 * time.Second
	}

	if cb.config.ReadyToTrip == nil {
		cb.config.ReadyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 5
		}
	}

	return cb
}

// Call executes the given function if the circuit breaker allows it.
// It tracks the success/failure and updates the circuit breaker state accordingly.
//
// Returns ErrCircuitOpen if the circuit breaker is open.
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.afterRequest(err == nil)

	return err
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current statistics.
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// beforeRequest checks if the request should be allowed.
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return ErrCircuitOpen
	}

	if state == StateHalfOpen && cb.counts.Requests >= cb.config.MaxRequests {
		return ErrCircuitOpen
	}

	cb.counts.Requests++
	cb.expiry = now.Add(cb.config.Interval)

	_ = generation // unused but part of state tracking

	return nil
}

// afterRequest records the result of a request.
func (cb *CircuitBreaker) afterRequest(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)

	if success {
		cb.onSuccess(state)
	} else {
		cb.onFailure(state)
	}
}

// onSuccess handles a successful request.
func (cb *CircuitBreaker) onSuccess(state State) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen {
		// Success in half-open state closes the circuit
		cb.setState(StateClosed)
	}
}

// onFailure handles a failed request.
func (cb *CircuitBreaker) onFailure(state State) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if state == StateHalfOpen {
		// Any failure in half-open state opens the circuit
		cb.setState(StateOpen)
	} else if cb.config.ReadyToTrip(cb.counts) {
		// In closed state, open if threshold is reached
		cb.setState(StateOpen)
	}
}

// currentState returns the current state and generation.
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		// Only reset counts if interval is configured and expired
		if cb.config.Interval > 0 && !cb.expiry.IsZero() && cb.expiry.Before(now) {
			// Reset counts after interval
			cb.counts = Counts{}
			cb.expiry = now.Add(cb.config.Interval)
		}
	case StateOpen:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			// Transition to half-open after timeout
			cb.setState(StateHalfOpen)
		}
	}

	return cb.state, 0
}

// setState transitions to a new state.
func (cb *CircuitBreaker) setState(state State) {
	if cb.state == state {
		return
	}

	prevState := cb.state
	cb.state = state

	// Reset counts when transitioning to closed or half-open
	if state == StateClosed || state == StateHalfOpen {
		cb.counts = Counts{}
	}

	if state == StateOpen {
		cb.expiry = time.Now().Add(cb.config.Timeout)
	} else {
		cb.expiry = time.Time{} // zero value
	}

	_ = prevState // for future logging
}

// Reset resets the circuit breaker to closed state with zero counts.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.counts = Counts{}
	cb.expiry = time.Time{}
}

// String returns a string representation of the circuit breaker state.
func (cb *CircuitBreaker) String() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return fmt.Sprintf("CircuitBreaker[state=%s, counts=%+v]", cb.state, cb.counts)
}
