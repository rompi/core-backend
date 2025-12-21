package httpclient

import "testing"

func TestNoopLogger(t *testing.T) {
	logger := NewNoopLogger()

	// All these should not panic
	logger.Debug("test", "key", "value")
	logger.Info("test", "key", "value")
	logger.Warn("test", "key", "value")
	logger.Error("test", "key", "value")
}
