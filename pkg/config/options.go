package config

import "time"

// Option configures the Config instance.
type Option func(*configImpl)

// WithProvider adds a configuration provider.
// Providers are loaded in order, with later providers overriding earlier ones.
func WithProvider(provider Provider) Option {
	return func(c *configImpl) {
		if provider != nil {
			c.providers = append(c.providers, provider)
		}
	}
}

// WithValidator sets a validator for configuration binding.
// The validator is called after binding to validate the struct.
func WithValidator(validator Validator) Option {
	return func(c *configImpl) {
		c.validator = validator
	}
}

// WithWatchInterval sets the interval for checking configuration changes.
// Default is 30 seconds.
func WithWatchInterval(interval time.Duration) Option {
	return func(c *configImpl) {
		if interval > 0 {
			c.watchInterval = interval
		}
	}
}

// WithKeyDelimiter sets the delimiter used for nested configuration keys.
// Default is "." (e.g., "database.host").
func WithKeyDelimiter(delimiter string) Option {
	return func(c *configImpl) {
		if delimiter != "" {
			c.keyDelimiter = delimiter
		}
	}
}

// WithSensitiveKey marks a key as sensitive.
// Sensitive keys are masked in AllSettings() output.
func WithSensitiveKey(key string) Option {
	return func(c *configImpl) {
		c.sensitiveKeys[key] = true
	}
}

// WithSensitiveKeys marks multiple keys as sensitive.
func WithSensitiveKeys(keys ...string) Option {
	return func(c *configImpl) {
		for _, key := range keys {
			c.sensitiveKeys[key] = true
		}
	}
}
