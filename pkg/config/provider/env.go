package provider

import (
	"context"
	"os"
	"strings"
)

// EnvOption configures the EnvProvider.
type EnvOption func(*EnvProvider)

// EnvProvider reads configuration from environment variables.
type EnvProvider struct {
	prefix    string
	delimiter string
}

// NewEnvProvider creates a new environment variable provider.
func NewEnvProvider(opts ...EnvOption) *EnvProvider {
	p := &EnvProvider{
		prefix:    "",
		delimiter: "_",
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithPrefix sets the prefix for environment variable names.
// For example, with prefix "APP", the key "database.host" would map to "APP_DATABASE_HOST".
func WithPrefix(prefix string) EnvOption {
	return func(p *EnvProvider) {
		p.prefix = strings.ToUpper(prefix)
	}
}

// WithDelimiter sets the delimiter used between key segments.
// Default is "_" (e.g., DATABASE_HOST).
func WithDelimiter(delimiter string) EnvOption {
	return func(p *EnvProvider) {
		p.delimiter = delimiter
	}
}

// Name returns the provider name.
func (p *EnvProvider) Name() string {
	return "env"
}

// Load loads configuration from environment variables.
func (p *EnvProvider) Load(_ context.Context) (map[string]any, error) {
	result := make(map[string]any)

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Filter by prefix if set
		if p.prefix != "" {
			if !strings.HasPrefix(key, p.prefix+p.delimiter) {
				continue
			}
			// Remove prefix
			key = strings.TrimPrefix(key, p.prefix+p.delimiter)
		}

		// Convert environment variable naming to dot notation
		// e.g., DATABASE_HOST -> database.host
		configKey := strings.ToLower(strings.ReplaceAll(key, p.delimiter, "."))

		result[configKey] = value
	}

	return result, nil
}

// Watch is not supported for environment variables.
// Environment variables are only read at load time.
func (p *EnvProvider) Watch(_ context.Context, _ func()) error {
	// Environment variables don't support watching
	return nil
}
