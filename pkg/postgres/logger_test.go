package postgres

import "testing"

func TestNewNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	if logger == nil {
		t.Fatal("NewNoopLogger() returned nil")
	}
}

func TestNoopLogger_Methods(t *testing.T) {
	logger := NewNoopLogger()

	// These should not panic
	logger.Debug("debug message")
	logger.Debug("debug with args", "key1", "value1", "key2", 123)
	logger.Info("info message")
	logger.Info("info with args", "key", "value")
	logger.Warn("warn message")
	logger.Warn("warn with args", "error", "something")
	logger.Error("error message")
	logger.Error("error with args", "detail", "failed")
}

func TestNoopLogger_ImplementsLogger(t *testing.T) {
	var _ Logger = (*NoopLogger)(nil)
	var _ Logger = NewNoopLogger()
}

func TestLoggerInterface(t *testing.T) {
	// Verify the interface contract
	var logger Logger = NewNoopLogger()

	// Should not panic with various argument types
	logger.Debug("msg", "string", "value", "int", 42, "float", 3.14, "bool", true)
	logger.Info("msg", "nil", nil, "slice", []string{"a", "b"})
	logger.Warn("msg")
	logger.Error("msg", "key", "value")
}
