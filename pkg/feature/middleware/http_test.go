package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rompi/core-backend/pkg/feature"
)

func TestHTTP_ContextInjection(t *testing.T) {
	client := &mockClient{}

	middleware := HTTP(client)

	var capturedCtx *feature.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = feature.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	if capturedCtx == nil {
		t.Fatal("Context should be injected")
	}
	if capturedCtx.Key != "user-123" {
		t.Errorf("Key = %v, want user-123", capturedCtx.Key)
	}
	if capturedCtx.IP != "192.168.1.1" {
		t.Errorf("IP = %v, want 192.168.1.1", capturedCtx.IP)
	}
}

func TestHTTP_CustomContextExtractor(t *testing.T) {
	client := &mockClient{}

	customExtractor := func(r *http.Request) *feature.Context {
		return feature.NewContext("custom-user").
			WithAttribute("custom", "value")
	}

	middleware := HTTP(client, WithContextExtractor(customExtractor))

	var capturedCtx *feature.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = feature.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	if capturedCtx == nil {
		t.Fatal("Context should be injected")
	}
	if capturedCtx.Key != "custom-user" {
		t.Errorf("Key = %v, want custom-user", capturedCtx.Key)
	}
	if capturedCtx.Custom["custom"] != "value" {
		t.Errorf("Custom[custom] = %v, want value", capturedCtx.Custom["custom"])
	}
}

func TestFeatureGate_Enabled(t *testing.T) {
	client := &mockClient{
		boolValues: map[string]bool{
			"new-feature": true,
		},
	}

	var handlerCalled bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	gate := FeatureGate(client, "new-feature", nil)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	gate(handler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Handler should be called when feature is enabled")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestFeatureGate_Disabled(t *testing.T) {
	client := &mockClient{
		boolValues: map[string]bool{
			"new-feature": false,
		},
	}

	var handlerCalled bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	gate := FeatureGate(client, "new-feature", nil)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	gate(handler).ServeHTTP(rr, req)

	if handlerCalled {
		t.Error("Handler should not be called when feature is disabled")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestFeatureGate_DisabledWithFallback(t *testing.T) {
	client := &mockClient{
		boolValues: map[string]bool{
			"new-feature": false,
		},
	}

	var fallbackCalled bool
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	gate := FeatureGate(client, "new-feature", fallback)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	gate(handler).ServeHTTP(rr, req)

	if !fallbackCalled {
		t.Error("Fallback should be called when feature is disabled")
	}
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
}

func TestVariantRouter(t *testing.T) {
	client := &mockClient{
		stringValues: map[string]string{
			"button-color": "blue",
		},
	}

	var calledHandler string
	blueHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledHandler = "blue"
		w.WriteHeader(http.StatusOK)
	})
	greenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledHandler = "green"
		w.WriteHeader(http.StatusOK)
	})

	router := NewVariantRouter(client, "button-color").
		Variant("blue", blueHandler).
		Variant("green", greenHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if calledHandler != "blue" {
		t.Errorf("Called handler = %v, want blue", calledHandler)
	}
}

func TestVariantRouter_Fallback(t *testing.T) {
	client := &mockClient{
		stringValues: map[string]string{
			"button-color": "unknown",
		},
	}

	var fallbackCalled bool
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusOK)
	})

	router := NewVariantRouter(client, "button-color").
		Variant("blue", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		Fallback(fallback)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !fallbackCalled {
		t.Error("Fallback should be called for unknown variant")
	}
}

func TestVariantRouter_NoFallback(t *testing.T) {
	client := &mockClient{
		stringValues: map[string]string{
			"button-color": "unknown",
		},
	}

	router := NewVariantRouter(client, "button-color").
		Variant("blue", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestVariantRouter_Handler(t *testing.T) {
	client := &mockClient{}
	router := NewVariantRouter(client, "test")

	handler := router.Handler()
	if handler == nil {
		t.Error("Handler() should return non-nil handler")
	}
}

// mockClient is a simple mock for testing middleware.
type mockClient struct {
	boolValues   map[string]bool
	stringValues map[string]string
}

func (m *mockClient) Bool(ctx context.Context, key string, defaultValue bool) bool {
	if v, ok := m.boolValues[key]; ok {
		return v
	}
	return defaultValue
}

func (m *mockClient) String(ctx context.Context, key string, defaultValue string) string {
	if v, ok := m.stringValues[key]; ok {
		return v
	}
	return defaultValue
}

func (m *mockClient) Int(ctx context.Context, key string, defaultValue int) int {
	return defaultValue
}

func (m *mockClient) Float(ctx context.Context, key string, defaultValue float64) float64 {
	return defaultValue
}

func (m *mockClient) JSON(ctx context.Context, key string, target interface{}) error {
	return nil
}

func (m *mockClient) Variation(ctx context.Context, key string) (*feature.Evaluation, error) {
	return &feature.Evaluation{Key: key}, nil
}

func (m *mockClient) AllFlags(ctx context.Context) map[string]interface{} {
	return make(map[string]interface{})
}

func (m *mockClient) Track(ctx context.Context, event string, data map[string]interface{}) {}

func (m *mockClient) Close() error {
	return nil
}
