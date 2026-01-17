// Package config provides unified configuration management supporting multiple
// sources including environment variables, files (YAML, JSON, TOML), and remote
// stores. It features type-safe configuration with validation, hot reloading,
// and struct tag-based binding.
package config

import "errors"

var (
	// ErrNotFound is returned when a configuration key does not exist.
	ErrNotFound = errors.New("config: key not found")

	// ErrInvalidType is returned when a value cannot be converted to the requested type.
	ErrInvalidType = errors.New("config: invalid type")

	// ErrRequired is returned when a required configuration field is missing.
	ErrRequired = errors.New("config: required field missing")

	// ErrValidation is returned when configuration validation fails.
	ErrValidation = errors.New("config: validation failed")

	// ErrProviderFailed is returned when a configuration provider fails to load.
	ErrProviderFailed = errors.New("config: provider failed")

	// ErrBindFailed is returned when binding configuration to a struct fails.
	ErrBindFailed = errors.New("config: bind failed")

	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("config: invalid configuration")

	// ErrParseError is returned when parsing configuration data fails.
	ErrParseError = errors.New("config: parse error")

	// ErrWatchFailed is returned when watching for configuration changes fails.
	ErrWatchFailed = errors.New("config: watch failed")
)

// BindError provides detailed information about a binding failure.
type BindError struct {
	Field   string // The field that failed to bind
	Tag     string // The struct tag being processed
	Value   any    // The value that caused the failure
	Message string // Human-readable error message
}

// Error implements the error interface.
func (e *BindError) Error() string {
	return e.Message
}

// MultiBindError contains multiple binding errors.
type MultiBindError struct {
	Errors []BindError
}

// Error implements the error interface.
func (e *MultiBindError) Error() string {
	if len(e.Errors) == 0 {
		return "config: no binding errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return "config: multiple binding errors occurred"
}

// Add appends a binding error.
func (e *MultiBindError) Add(err BindError) {
	e.Errors = append(e.Errors, err)
}

// HasErrors returns true if there are any binding errors.
func (e *MultiBindError) HasErrors() bool {
	return len(e.Errors) > 0
}
