package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// CheckerFunc is a function that performs a health check.
type CheckerFunc func(ctx context.Context) error

// Checker manages health checks.
type Checker struct {
	checks map[string]CheckerFunc
	mu     sync.RWMutex
}

// NewChecker creates a new health checker.
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]CheckerFunc),
	}
}

// Add adds a named health check.
func (c *Checker) Add(name string, check CheckerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// Remove removes a health check by name.
func (c *Checker) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// Check runs all health checks and returns the status.
func (c *Checker) Check(ctx context.Context) *Status {
	c.mu.RLock()
	checks := make(map[string]CheckerFunc, len(c.checks))
	for name, fn := range c.checks {
		checks[name] = fn
	}
	c.mu.RUnlock()

	start := time.Now()
	results := make(map[string]CheckResult)
	status := StatusHealthy

	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, fn := range checks {
		wg.Add(1)
		go func(name string, fn CheckerFunc) {
			defer wg.Done()

			checkStart := time.Now()
			err := fn(ctx)
			duration := time.Since(checkStart)

			result := CheckResult{
				Status:   StatusHealthy,
				Duration: duration.String(),
			}

			if err != nil {
				result.Status = StatusUnhealthy
				result.Error = err.Error()
			}

			mu.Lock()
			results[name] = result
			if result.Status == StatusUnhealthy {
				status = StatusUnhealthy
			}
			mu.Unlock()
		}(name, fn)
	}

	wg.Wait()

	return &Status{
		Status:    status,
		Timestamp: time.Now(),
		Duration:  time.Since(start).String(),
		Checks:    results,
	}
}

// IsHealthy returns true if all checks pass.
func (c *Checker) IsHealthy(ctx context.Context) bool {
	status := c.Check(ctx)
	return status.Status == StatusHealthy
}

// Status represents the health check response.
type Status struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  string                 `json:"duration"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult represents the result of a single health check.
type CheckResult struct {
	Status   string `json:"status"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

// Health status constants.
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusDegraded  = "degraded"
)

// --- HTTP Handlers ---

// HealthHandler creates an HTTP handler for health checks.
func HealthHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		status := checker.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		if status.Status != StatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}

// LivenessHandler creates an HTTP handler for liveness probes.
// Liveness indicates if the application is running.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": StatusHealthy,
		})
	}
}

// ReadinessHandler creates an HTTP handler for readiness probes.
// Readiness indicates if the application is ready to serve traffic.
func ReadinessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		status := checker.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		if status.Status != StatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}

// SimpleHandler creates a simple health handler that returns 200 OK.
func SimpleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
