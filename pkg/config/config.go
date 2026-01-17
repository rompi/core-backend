package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config manages configuration loading, binding, and access from multiple providers.
type Config interface {
	// Load loads configuration from all providers.
	Load(ctx context.Context) error

	// Bind binds configuration to a struct using struct tags.
	Bind(v any) error

	// Get returns a value by key, or nil if not found.
	Get(key string) any

	// GetString returns a string value for the key.
	GetString(key string) string

	// GetInt returns an int value for the key.
	GetInt(key string) int

	// GetInt64 returns an int64 value for the key.
	GetInt64(key string) int64

	// GetFloat64 returns a float64 value for the key.
	GetFloat64(key string) float64

	// GetBool returns a bool value for the key.
	GetBool(key string) bool

	// GetDuration returns a time.Duration value for the key.
	GetDuration(key string) time.Duration

	// GetTime returns a time.Time value for the key.
	GetTime(key string) time.Time

	// GetStringSlice returns a string slice for the key.
	GetStringSlice(key string) []string

	// GetStringMap returns a string map for the key.
	GetStringMap(key string) map[string]any

	// IsSet returns true if the key is set.
	IsSet(key string) bool

	// Set sets a value for the key.
	Set(key string, value any)

	// Watch watches for configuration changes and calls the callback on changes.
	Watch(ctx context.Context, callback func(Config)) error

	// Sub returns a sub-configuration for the given key prefix.
	Sub(key string) Config

	// AllSettings returns all settings as a map.
	AllSettings() map[string]any
}

// Provider provides configuration from a source.
type Provider interface {
	// Name returns the provider name for identification.
	Name() string

	// Load loads configuration from the source.
	Load(ctx context.Context) (map[string]any, error)

	// Watch watches for changes and calls the callback when changes occur.
	// Returns nil if watching is not supported.
	Watch(ctx context.Context, callback func()) error
}

// Parser parses configuration data from bytes.
type Parser interface {
	// Parse parses data into a configuration map.
	Parse(data []byte) (map[string]any, error)

	// Extensions returns the file extensions this parser supports.
	Extensions() []string
}

// Validator validates configuration values.
type Validator interface {
	// Validate validates a struct and returns an error if invalid.
	Validate(v any) error
}

// configImpl is the default implementation of Config.
type configImpl struct {
	mu            sync.RWMutex
	providers     []Provider
	values        map[string]any
	validator     Validator
	watchInterval time.Duration
	keyDelimiter  string
	sensitiveKeys map[string]bool
}

// New creates a new Config instance with the provided options.
func New(opts ...Option) Config {
	c := &configImpl{
		providers:     []Provider{},
		values:        make(map[string]any),
		watchInterval: 30 * time.Second,
		keyDelimiter:  ".",
		sensitiveKeys: make(map[string]bool),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Load loads configuration from all providers in order.
// Later providers override earlier ones.
func (c *configImpl) Load(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear existing values
	c.values = make(map[string]any)

	// Load from each provider
	for _, provider := range c.providers {
		values, err := provider.Load(ctx)
		if err != nil {
			return fmt.Errorf("%w: provider %s: %v", ErrProviderFailed, provider.Name(), err)
		}

		// Merge values (later providers override)
		c.mergeValues(values, "")
	}

	return nil
}

// mergeValues merges nested values into the flat key-value store.
func (c *configImpl) mergeValues(values map[string]any, prefix string) {
	for k, v := range values {
		key := k
		if prefix != "" {
			key = prefix + c.keyDelimiter + k
		}

		// Recursively handle nested maps
		if nested, ok := v.(map[string]any); ok {
			c.mergeValues(nested, key)
		} else {
			c.values[strings.ToLower(key)] = v
		}
	}
}

// Bind binds configuration to a struct.
func (c *configImpl) Bind(v any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	binder := newBinder(c.values, c.keyDelimiter)
	if err := binder.Bind(v); err != nil {
		return fmt.Errorf("%w: %v", ErrBindFailed, err)
	}

	// Run validation if validator is set
	if c.validator != nil {
		if err := c.validator.Validate(v); err != nil {
			return fmt.Errorf("%w: %v", ErrValidation, err)
		}
	}

	return nil
}

// Get returns a value by key.
func (c *configImpl) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.values[strings.ToLower(key)]
}

// GetString returns a string value.
func (c *configImpl) GetString(key string) string {
	v := c.Get(key)
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt returns an int value.
func (c *configImpl) GetInt(key string) int {
	return int(c.GetInt64(key))
}

// GetInt64 returns an int64 value.
func (c *configImpl) GetInt64(key string) int64 {
	v := c.Get(key)
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}

// GetFloat64 returns a float64 value.
func (c *configImpl) GetFloat64(key string) float64 {
	v := c.Get(key)
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// GetBool returns a bool value.
func (c *configImpl) GetBool(key string) bool {
	v := c.Get(key)
	if v == nil {
		return false
	}

	switch val := v.(type) {
	case bool:
		return val
	case string:
		return parseBool(val)
	case int:
		return val != 0
	case int64:
		return val != 0
	default:
		return false
	}
}

// GetDuration returns a time.Duration value.
func (c *configImpl) GetDuration(key string) time.Duration {
	v := c.Get(key)
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case time.Duration:
		return val
	case string:
		d, _ := time.ParseDuration(val)
		return d
	case int:
		return time.Duration(val)
	case int64:
		return time.Duration(val)
	case float64:
		return time.Duration(val)
	default:
		return 0
	}
}

// GetTime returns a time.Time value.
func (c *configImpl) GetTime(key string) time.Time {
	v := c.Get(key)
	if v == nil {
		return time.Time{}
	}

	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		// Try common formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",
			"2006-01-02 15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t
			}
		}
		return time.Time{}
	default:
		return time.Time{}
	}
}

// GetStringSlice returns a string slice.
func (c *configImpl) GetStringSlice(key string) []string {
	v := c.Get(key)
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case string:
		// Split comma-separated values
		if val == "" {
			return nil
		}
		parts := strings.Split(val, ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result
	default:
		return nil
	}
}

// GetStringMap returns a string map.
func (c *configImpl) GetStringMap(key string) map[string]any {
	v := c.Get(key)
	if v == nil {
		// Try to build map from prefixed keys
		return c.buildNestedMap(key)
	}

	if m, ok := v.(map[string]any); ok {
		return m
	}

	return nil
}

// buildNestedMap builds a nested map from keys with the given prefix.
func (c *configImpl) buildNestedMap(prefix string) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any)
	prefix = strings.ToLower(prefix) + c.keyDelimiter

	for k, v := range c.values {
		if strings.HasPrefix(k, prefix) {
			subKey := strings.TrimPrefix(k, prefix)
			result[subKey] = v
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// IsSet returns true if the key is set.
func (c *configImpl) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.values[strings.ToLower(key)]
	return ok
}

// Set sets a value for the key.
func (c *configImpl) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[strings.ToLower(key)] = value
}

// Watch watches for configuration changes.
func (c *configImpl) Watch(ctx context.Context, callback func(Config)) error {
	// Start watching each provider that supports it
	for _, provider := range c.providers {
		providerCallback := func() {
			// Reload configuration
			if err := c.Load(ctx); err == nil {
				callback(c)
			}
		}

		if err := provider.Watch(ctx, providerCallback); err != nil {
			// Not all providers support watching, ignore errors
			continue
		}
	}

	return nil
}

// Sub returns a sub-configuration.
func (c *configImpl) Sub(key string) Config {
	return &subConfig{
		parent: c,
		prefix: key,
	}
}

// AllSettings returns all settings as a map.
func (c *configImpl) AllSettings() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to avoid data races
	result := make(map[string]any, len(c.values))
	for k, v := range c.values {
		// Mask sensitive values
		if c.sensitiveKeys[k] {
			result[k] = "***"
		} else {
			result[k] = v
		}
	}

	return result
}

// subConfig wraps a Config with a key prefix.
type subConfig struct {
	parent Config
	prefix string
}

func (s *subConfig) Load(ctx context.Context) error {
	return s.parent.Load(ctx)
}

func (s *subConfig) Bind(v any) error {
	return s.parent.Bind(v)
}

func (s *subConfig) Get(key string) any {
	return s.parent.Get(s.prefix + "." + key)
}

func (s *subConfig) GetString(key string) string {
	return s.parent.GetString(s.prefix + "." + key)
}

func (s *subConfig) GetInt(key string) int {
	return s.parent.GetInt(s.prefix + "." + key)
}

func (s *subConfig) GetInt64(key string) int64 {
	return s.parent.GetInt64(s.prefix + "." + key)
}

func (s *subConfig) GetFloat64(key string) float64 {
	return s.parent.GetFloat64(s.prefix + "." + key)
}

func (s *subConfig) GetBool(key string) bool {
	return s.parent.GetBool(s.prefix + "." + key)
}

func (s *subConfig) GetDuration(key string) time.Duration {
	return s.parent.GetDuration(s.prefix + "." + key)
}

func (s *subConfig) GetTime(key string) time.Time {
	return s.parent.GetTime(s.prefix + "." + key)
}

func (s *subConfig) GetStringSlice(key string) []string {
	return s.parent.GetStringSlice(s.prefix + "." + key)
}

func (s *subConfig) GetStringMap(key string) map[string]any {
	return s.parent.GetStringMap(s.prefix + "." + key)
}

func (s *subConfig) IsSet(key string) bool {
	return s.parent.IsSet(s.prefix + "." + key)
}

func (s *subConfig) Set(key string, value any) {
	s.parent.Set(s.prefix+"."+key, value)
}

func (s *subConfig) Watch(ctx context.Context, callback func(Config)) error {
	return s.parent.Watch(ctx, callback)
}

func (s *subConfig) Sub(key string) Config {
	return &subConfig{
		parent: s.parent,
		prefix: s.prefix + "." + key,
	}
}

func (s *subConfig) AllSettings() map[string]any {
	// Get all settings and filter by prefix
	all := s.parent.AllSettings()
	result := make(map[string]any)
	prefix := strings.ToLower(s.prefix) + "."

	for k, v := range all {
		if strings.HasPrefix(k, prefix) {
			subKey := strings.TrimPrefix(k, prefix)
			result[subKey] = v
		}
	}

	return result
}

// parseBool parses a string as a boolean value.
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}
