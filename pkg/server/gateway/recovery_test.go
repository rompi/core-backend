package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/rompi/core-backend/pkg/server"
)

// mockLogger captures log messages for testing
type mockLogger struct {
	mu       sync.Mutex
	messages []logMessage
}

type logMessage struct {
	level  string
	msg    string
	fields []any
}

func (l *mockLogger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "debug", msg: msg, fields: args})
}

func (l *mockLogger) Info(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "info", msg: msg, fields: args})
}

func (l *mockLogger) Warn(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "warn", msg: msg, fields: args})
}

func (l *mockLogger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "error", msg: msg, fields: args})
}

func TestDefaultRecoveryConfig(t *testing.T) {
	cfg := DefaultRecoveryConfig()

	if cfg.Logger == nil {
		t.Error("Logger should not be nil")
	}

	if !cfg.EnableStack {
		t.Error("EnableStack should be true by default")
	}

	if cfg.ErrorHandler != nil {
		t.Error("ErrorHandler should be nil by default")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := &mockLogger{}
	middleware := RecoveryMiddleware(logger)

	t.Run("normal request passes through", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if rr.Body.String() != "success" {
			t.Errorf("body = %q, want %q", rr.Body.String(), "success")
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		logger.messages = nil // Reset logger

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		}))

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rr := httptest.NewRecorder()

		// Should not panic
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}

		if !strings.Contains(rr.Body.String(), "Internal Server Error") {
			t.Errorf("body = %q, want Internal Server Error", rr.Body.String())
		}

		// Check that error was logged
		if len(logger.messages) == 0 {
			t.Error("panic should have been logged")
		}

		found := false
		for _, m := range logger.messages {
			if m.level == "error" && m.msg == "panic recovered" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'panic recovered' error log")
		}
	})
}

func TestRecoveryMiddlewareWithConfig(t *testing.T) {
	t.Run("with nil logger uses noop", func(t *testing.T) {
		config := RecoveryConfig{
			Logger:      nil,
			EnableStack: false,
		}
		middleware := RecoveryMiddlewareWithConfig(config)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Should not panic
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("with stack trace disabled", func(t *testing.T) {
		logger := &mockLogger{}
		config := RecoveryConfig{
			Logger:      logger,
			EnableStack: false,
		}
		middleware := RecoveryMiddlewareWithConfig(config)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Check that stack trace was not included
		for _, m := range logger.messages {
			for i := 0; i < len(m.fields)-1; i += 2 {
				if m.fields[i] == "stack" {
					t.Error("stack should not be logged when EnableStack is false")
				}
			}
		}
	})

	t.Run("with stack trace enabled", func(t *testing.T) {
		logger := &mockLogger{}
		config := RecoveryConfig{
			Logger:      logger,
			EnableStack: true,
		}
		middleware := RecoveryMiddlewareWithConfig(config)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Check that stack trace was included
		hasStack := false
		for _, m := range logger.messages {
			for i := 0; i < len(m.fields)-1; i += 2 {
				if m.fields[i] == "stack" {
					hasStack = true
					break
				}
			}
		}
		if !hasStack {
			t.Error("stack should be logged when EnableStack is true")
		}
	})

	t.Run("with custom error handler", func(t *testing.T) {
		customHandlerCalled := false
		config := RecoveryConfig{
			Logger:      server.NoopLogger{},
			EnableStack: false,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err any) {
				customHandlerCalled = true
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("custom error"))
			},
		}
		middleware := RecoveryMiddlewareWithConfig(config)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if !customHandlerCalled {
			t.Error("custom error handler should have been called")
		}
		if rr.Code != http.StatusTeapot {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusTeapot)
		}
		if rr.Body.String() != "custom error" {
			t.Errorf("body = %q, want %q", rr.Body.String(), "custom error")
		}
	})
}

func TestRecoveryMiddleware_PanicTypes(t *testing.T) {
	logger := &mockLogger{}
	middleware := RecoveryMiddleware(logger)

	tests := []struct {
		name       string
		panicValue any
	}{
		{"string panic", "string error"},
		{"error panic", http.ErrAbortHandler},
		{"int panic", 42},
		{"struct panic", struct{ msg string }{"error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.messages = nil

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicValue)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
			}
		})
	}
}

func TestRecoveryMiddleware_LogsRequestInfo(t *testing.T) {
	logger := &mockLogger{}
	config := RecoveryConfig{
		Logger:      logger,
		EnableStack: false,
	}
	middleware := RecoveryMiddlewareWithConfig(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Verify method and path were logged
	if len(logger.messages) == 0 {
		t.Fatal("expected log messages")
	}

	msg := logger.messages[0]
	hasMethod := false
	hasPath := false

	for i := 0; i < len(msg.fields)-1; i += 2 {
		key, ok := msg.fields[i].(string)
		if !ok {
			continue
		}
		if key == "method" {
			if val, ok := msg.fields[i+1].(string); ok && val == "POST" {
				hasMethod = true
			}
		}
		if key == "path" {
			if val, ok := msg.fields[i+1].(string); ok && val == "/api/test" {
				hasPath = true
			}
		}
	}

	if !hasMethod {
		t.Error("log should include method")
	}
	if !hasPath {
		t.Error("log should include path")
	}
}
