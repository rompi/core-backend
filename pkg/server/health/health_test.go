package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	checker := NewChecker()
	if checker == nil {
		t.Fatal("NewChecker() returned nil")
	}
	if checker.checks == nil {
		t.Error("NewChecker() checks map is nil")
	}
}

func TestChecker_Add(t *testing.T) {
	checker := NewChecker()

	checker.Add("db", func(ctx context.Context) error {
		return nil
	})

	if _, ok := checker.checks["db"]; !ok {
		t.Error("Add() did not add check")
	}
}

func TestChecker_Remove(t *testing.T) {
	checker := NewChecker()

	checker.Add("db", func(ctx context.Context) error {
		return nil
	})
	checker.Remove("db")

	if _, ok := checker.checks["db"]; ok {
		t.Error("Remove() did not remove check")
	}
}

func TestChecker_Check_AllHealthy(t *testing.T) {
	checker := NewChecker()

	checker.Add("db", func(ctx context.Context) error {
		return nil
	})
	checker.Add("cache", func(ctx context.Context) error {
		return nil
	})

	status := checker.Check(context.Background())

	if status.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", status.Status, StatusHealthy)
	}
	if len(status.Checks) != 2 {
		t.Errorf("Checks len = %d, want 2", len(status.Checks))
	}
	if status.Checks["db"].Status != StatusHealthy {
		t.Error("db check should be healthy")
	}
	if status.Checks["cache"].Status != StatusHealthy {
		t.Error("cache check should be healthy")
	}
	if status.Duration == "" {
		t.Error("Duration should not be empty")
	}
	if status.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestChecker_Check_OneUnhealthy(t *testing.T) {
	checker := NewChecker()

	checker.Add("db", func(ctx context.Context) error {
		return nil
	})
	checker.Add("cache", func(ctx context.Context) error {
		return errors.New("connection failed")
	})

	status := checker.Check(context.Background())

	if status.Status != StatusUnhealthy {
		t.Errorf("Status = %q, want %q", status.Status, StatusUnhealthy)
	}
	if status.Checks["db"].Status != StatusHealthy {
		t.Error("db check should be healthy")
	}
	if status.Checks["cache"].Status != StatusUnhealthy {
		t.Error("cache check should be unhealthy")
	}
	if status.Checks["cache"].Error != "connection failed" {
		t.Errorf("cache error = %q, want %q", status.Checks["cache"].Error, "connection failed")
	}
}

func TestChecker_Check_AllUnhealthy(t *testing.T) {
	checker := NewChecker()

	checker.Add("db", func(ctx context.Context) error {
		return errors.New("db error")
	})
	checker.Add("cache", func(ctx context.Context) error {
		return errors.New("cache error")
	})

	status := checker.Check(context.Background())

	if status.Status != StatusHealthy {
		// Both should fail individually, but the overall status should be unhealthy
		// Wait - let me re-check the logic...
	}
	if status.Status != StatusUnhealthy {
		t.Errorf("Status = %q, want %q", status.Status, StatusUnhealthy)
	}
}

func TestChecker_Check_NoChecks(t *testing.T) {
	checker := NewChecker()

	status := checker.Check(context.Background())

	if status.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q for no checks", status.Status, StatusHealthy)
	}
	if len(status.Checks) != 0 {
		t.Errorf("Checks len = %d, want 0", len(status.Checks))
	}
}

func TestChecker_Check_Concurrent(t *testing.T) {
	checker := NewChecker()

	var callCount int32

	checker.Add("check1", func(ctx context.Context) error {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	checker.Add("check2", func(ctx context.Context) error {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	start := time.Now()
	status := checker.Check(context.Background())
	elapsed := time.Since(start)

	if status.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", status.Status, StatusHealthy)
	}
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
	// If run concurrently, should be around 10ms, not 20ms
	if elapsed > 50*time.Millisecond {
		t.Errorf("elapsed = %v, checks should run concurrently", elapsed)
	}
}

func TestChecker_IsHealthy(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return nil
		})

		if !checker.IsHealthy(context.Background()) {
			t.Error("IsHealthy() = false, want true")
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return errors.New("error")
		})

		if checker.IsHealthy(context.Background()) {
			t.Error("IsHealthy() = true, want false")
		}
	})
}

func TestHealthHandler(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return nil
		})

		handler := HealthHandler(checker)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		var status Status
		if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if status.Status != StatusHealthy {
			t.Errorf("status = %q, want %q", status.Status, StatusHealthy)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return errors.New("connection failed")
		})

		handler := HealthHandler(checker)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusServiceUnavailable)
		}

		var status Status
		if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if status.Status != StatusUnhealthy {
			t.Errorf("status = %q, want %q", status.Status, StatusUnhealthy)
		}
	})
}

func TestLivenessHandler(t *testing.T) {
	handler := LivenessHandler()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != StatusHealthy {
		t.Errorf("status = %q, want %q", response["status"], StatusHealthy)
	}
}

func TestReadinessHandler(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return nil
		})

		handler := ReadinessHandler(checker)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
		}

		var status Status
		if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if status.Status != StatusHealthy {
			t.Errorf("status = %q, want %q", status.Status, StatusHealthy)
		}
	})

	t.Run("not ready", func(t *testing.T) {
		checker := NewChecker()
		checker.Add("db", func(ctx context.Context) error {
			return errors.New("not ready")
		})

		handler := ReadinessHandler(checker)
		req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("status code = %d, want %d", rr.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestSimpleHandler(t *testing.T) {
	handler := SimpleHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "OK")
	}
}

func TestChecker_ThreadSafety(t *testing.T) {
	checker := NewChecker()

	// Concurrent add/remove/check operations
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			checker.Add("check", func(ctx context.Context) error {
				return nil
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			checker.Remove("check")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			checker.Check(context.Background())
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusHealthy != "healthy" {
		t.Errorf("StatusHealthy = %q, want %q", StatusHealthy, "healthy")
	}
	if StatusUnhealthy != "unhealthy" {
		t.Errorf("StatusUnhealthy = %q, want %q", StatusUnhealthy, "unhealthy")
	}
	if StatusDegraded != "degraded" {
		t.Errorf("StatusDegraded = %q, want %q", StatusDegraded, "degraded")
	}
}
