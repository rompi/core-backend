package feature

import "errors"

// Sentinel errors for common failure scenarios.
var (
	// ErrFlagNotFound is returned when a flag key does not exist.
	ErrFlagNotFound = errors.New("feature: flag not found")

	// ErrFlagDisabled is returned when a flag is disabled.
	ErrFlagDisabled = errors.New("feature: flag disabled")

	// ErrInvalidFlagType is returned when flag value type doesn't match expected type.
	ErrInvalidFlagType = errors.New("feature: invalid flag type")

	// ErrInvalidConfig is returned when configuration is invalid.
	ErrInvalidConfig = errors.New("feature: invalid configuration")

	// ErrProviderNotInitialized is returned when the provider is not initialized.
	ErrProviderNotInitialized = errors.New("feature: provider not initialized")

	// ErrInvalidContext is returned when evaluation context is invalid.
	ErrInvalidContext = errors.New("feature: invalid context")

	// ErrPrerequisiteFailed is returned when a prerequisite flag evaluation fails.
	ErrPrerequisiteFailed = errors.New("feature: prerequisite failed")

	// ErrFileNotFound is returned when the feature flags file is not found.
	ErrFileNotFound = errors.New("feature: file not found")

	// ErrInvalidFileFormat is returned when the file format is invalid.
	ErrInvalidFileFormat = errors.New("feature: invalid file format")
)
