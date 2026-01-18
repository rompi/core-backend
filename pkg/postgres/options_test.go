package postgres

import (
	"sync"
	"testing"
)

// mockLogger is a test logger that records calls
type mockLogger struct {
	mu       sync.Mutex
	messages []logMessage
}

type logMessage struct {
	level  string
	msg    string
	fields []any
}

func (l *mockLogger) Debug(msg string, keysAndValues ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "debug", msg: msg, fields: keysAndValues})
}

func (l *mockLogger) Info(msg string, keysAndValues ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "info", msg: msg, fields: keysAndValues})
}

func (l *mockLogger) Warn(msg string, keysAndValues ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "warn", msg: msg, fields: keysAndValues})
}

func (l *mockLogger) Error(msg string, keysAndValues ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, logMessage{level: "error", msg: msg, fields: keysAndValues})
}

func TestWithLogger(t *testing.T) {
	t.Run("sets logger", func(t *testing.T) {
		client := &Client{}
		logger := &mockLogger{}

		opt := WithLogger(logger)
		opt(client)

		if client.logger == nil {
			t.Error("logger should be set")
		}
	})

	t.Run("nil logger is ignored", func(t *testing.T) {
		client := &Client{logger: NewNoopLogger()}

		opt := WithLogger(nil)
		opt(client)

		// Logger should remain the original
		if client.logger == nil {
			t.Error("logger should not be nil")
		}
	})
}

// mockQueryHook is a test query hook
type mockQueryHook struct {
	beforeCalls []beforeCall
	afterCalls  []afterCall
	mu          sync.Mutex
}

type beforeCall struct {
	sql  string
	args []any
}

type afterCall struct {
	sql  string
	args []any
	err  error
}

func (h *mockQueryHook) BeforeQuery(sql string, args []any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.beforeCalls = append(h.beforeCalls, beforeCall{sql: sql, args: args})
}

func (h *mockQueryHook) AfterQuery(sql string, args []any, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.afterCalls = append(h.afterCalls, afterCall{sql: sql, args: args, err: err})
}

func TestWithQueryHook(t *testing.T) {
	client := &Client{}
	hook := &mockQueryHook{}

	opt := WithQueryHook(hook)
	opt(client)

	if client.queryHook == nil {
		t.Error("queryHook should be set")
	}
}

func TestQueryHookInterface(t *testing.T) {
	hook := &mockQueryHook{}

	// Test BeforeQuery
	hook.BeforeQuery("SELECT * FROM users WHERE id = $1", []any{1})
	if len(hook.beforeCalls) != 1 {
		t.Errorf("beforeCalls len = %d, want 1", len(hook.beforeCalls))
	}
	if hook.beforeCalls[0].sql != "SELECT * FROM users WHERE id = $1" {
		t.Errorf("sql = %q", hook.beforeCalls[0].sql)
	}

	// Test AfterQuery
	hook.AfterQuery("SELECT * FROM users WHERE id = $1", []any{1}, nil)
	if len(hook.afterCalls) != 1 {
		t.Errorf("afterCalls len = %d, want 1", len(hook.afterCalls))
	}
}
