package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthBearerMiddleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			t.Errorf("expected Authorization header 'Bearer test-token-123', got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(AuthBearerMiddleware("test-token-123"))

	_, err := client.Get(context.Background(), "/test").Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthAPIKeyMiddleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "secret-key-456" {
			t.Errorf("expected X-API-Key header 'secret-key-456', got %q", apiKey)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(AuthAPIKeyMiddleware("X-API-Key", "secret-key-456"))

	_, err := client.Get(context.Background(), "/test").Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserAgentMiddleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "CustomApp/1.0" {
			t.Errorf("expected User-Agent 'CustomApp/1.0', got %q", ua)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(UserAgentMiddleware("CustomApp/1.0"))

	_, err := client.Get(context.Background(), "/test").Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeaderMiddleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-1") != "value1" {
			t.Errorf("expected X-Custom-1 'value1', got %q", r.Header.Get("X-Custom-1"))
		}
		if r.Header.Get("X-Custom-2") != "value2" {
			t.Errorf("expected X-Custom-2 'value2', got %q", r.Header.Get("X-Custom-2"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(HeaderMiddleware(map[string]string{
		"X-Custom-1": "value1",
		"X-Custom-2": "value2",
	}))

	_, err := client.Get(context.Background(), "/test").Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logged := false
	logger := &testLogger{
		onInfo: func(msg string, keysAndValues ...interface{}) {
			logged = true
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(LoggingMiddleware(logger))

	_, err := client.Get(context.Background(), "/test").Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !logged {
		t.Error("expected logging middleware to log request")
	}
}

type testLogger struct {
	onDebug func(msg string, keysAndValues ...interface{})
	onInfo  func(msg string, keysAndValues ...interface{})
	onWarn  func(msg string, keysAndValues ...interface{})
	onError func(msg string, keysAndValues ...interface{})
}

func (l *testLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.onDebug != nil {
		l.onDebug(msg, keysAndValues...)
	}
}

func (l *testLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.onInfo != nil {
		l.onInfo(msg, keysAndValues...)
	}
}

func (l *testLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.onWarn != nil {
		l.onWarn(msg, keysAndValues...)
	}
}

func (l *testLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.onError != nil {
		l.onError(msg, keysAndValues...)
	}
}
