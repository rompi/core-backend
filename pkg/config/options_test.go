package config

import (
	"testing"
	"time"
)

func TestWithProvider(t *testing.T) {
	provider := &mockProvider{name: "test"}
	cfg := New(WithProvider(provider))

	impl := cfg.(*configImpl)
	if len(impl.providers) != 1 {
		t.Errorf("WithProvider() did not add provider, len = %d", len(impl.providers))
	}
}

func TestWithProvider_Nil(t *testing.T) {
	cfg := New(WithProvider(nil))

	impl := cfg.(*configImpl)
	if len(impl.providers) != 0 {
		t.Errorf("WithProvider(nil) added provider, len = %d", len(impl.providers))
	}
}

func TestWithValidator(t *testing.T) {
	validator := &mockValidator{}
	cfg := New(WithValidator(validator))

	impl := cfg.(*configImpl)
	if impl.validator != validator {
		t.Error("WithValidator() did not set validator")
	}
}

func TestWithWatchInterval(t *testing.T) {
	cfg := New(WithWatchInterval(10 * time.Second))

	impl := cfg.(*configImpl)
	if impl.watchInterval != 10*time.Second {
		t.Errorf("WithWatchInterval() = %v, want %v", impl.watchInterval, 10*time.Second)
	}
}

func TestWithWatchInterval_Zero(t *testing.T) {
	cfg := New(WithWatchInterval(0))

	impl := cfg.(*configImpl)
	// Should keep default value
	if impl.watchInterval != 30*time.Second {
		t.Errorf("WithWatchInterval(0) changed interval, got %v", impl.watchInterval)
	}
}

func TestWithKeyDelimiter(t *testing.T) {
	cfg := New(WithKeyDelimiter("_"))

	impl := cfg.(*configImpl)
	if impl.keyDelimiter != "_" {
		t.Errorf("WithKeyDelimiter() = %v, want %v", impl.keyDelimiter, "_")
	}
}

func TestWithKeyDelimiter_Empty(t *testing.T) {
	cfg := New(WithKeyDelimiter(""))

	impl := cfg.(*configImpl)
	// Should keep default value
	if impl.keyDelimiter != "." {
		t.Errorf("WithKeyDelimiter('') changed delimiter, got %v", impl.keyDelimiter)
	}
}

func TestWithSensitiveKey(t *testing.T) {
	cfg := New(WithSensitiveKey("password"))

	impl := cfg.(*configImpl)
	if !impl.sensitiveKeys["password"] {
		t.Error("WithSensitiveKey() did not mark key as sensitive")
	}
}

func TestWithSensitiveKeys(t *testing.T) {
	cfg := New(WithSensitiveKeys("password", "api_key", "token"))

	impl := cfg.(*configImpl)
	if !impl.sensitiveKeys["password"] {
		t.Error("WithSensitiveKeys() did not mark password as sensitive")
	}
	if !impl.sensitiveKeys["api_key"] {
		t.Error("WithSensitiveKeys() did not mark api_key as sensitive")
	}
	if !impl.sensitiveKeys["token"] {
		t.Error("WithSensitiveKeys() did not mark token as sensitive")
	}
}

// mockValidator for testing
type mockValidator struct {
	validateErr error
}

func (v *mockValidator) Validate(_ any) error {
	return v.validateErr
}
