package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// Client evaluates feature flags.
type Client interface {
	// Bool evaluates a boolean flag.
	Bool(ctx context.Context, key string, defaultValue bool) bool

	// String evaluates a string flag.
	String(ctx context.Context, key string, defaultValue string) string

	// Int evaluates an integer flag.
	Int(ctx context.Context, key string, defaultValue int) int

	// Float evaluates a float flag.
	Float(ctx context.Context, key string, defaultValue float64) float64

	// JSON evaluates a JSON flag into target.
	JSON(ctx context.Context, key string, target interface{}) error

	// Variation evaluates a flag with full details.
	Variation(ctx context.Context, key string) (*Evaluation, error)

	// AllFlags returns all flag values for a context.
	AllFlags(ctx context.Context) map[string]interface{}

	// Track records a custom event for analytics.
	Track(ctx context.Context, event string, data map[string]interface{})

	// Close releases resources.
	Close() error
}

// Provider defines the interface for feature flag backends.
type Provider interface {
	// GetFlag retrieves a flag by key.
	GetFlag(ctx context.Context, key string) (*Flag, error)

	// GetAllFlags retrieves all flags.
	GetAllFlags(ctx context.Context) (map[string]*Flag, error)

	// SetFlag creates or updates a flag.
	SetFlag(ctx context.Context, flag *Flag) error

	// DeleteFlag removes a flag.
	DeleteFlag(ctx context.Context, key string) error

	// Close releases provider resources.
	Close() error
}

// client is the default implementation of Client.
type client struct {
	mu       sync.RWMutex
	provider Provider
	options  *clientOptions
	closed   bool
}

// New creates a new feature flag client with the provided configuration.
func New(cfg Config, opts ...Option) (Client, error) {
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	c := &client{
		options: options,
	}

	return c, nil
}

// NewWithProvider creates a new client with a custom provider.
func NewWithProvider(provider Provider, opts ...Option) Client {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &client{
		provider: provider,
		options:  options,
	}
}

// Bool evaluates a boolean flag.
func (c *client) Bool(ctx context.Context, key string, defaultValue bool) bool {
	eval, err := c.evaluate(ctx, key, defaultValue)
	if err != nil {
		return defaultValue
	}

	if val, ok := eval.Value.(bool); ok {
		return val
	}

	return defaultValue
}

// String evaluates a string flag.
func (c *client) String(ctx context.Context, key string, defaultValue string) string {
	eval, err := c.evaluate(ctx, key, defaultValue)
	if err != nil {
		return defaultValue
	}

	if val, ok := eval.Value.(string); ok {
		return val
	}

	return defaultValue
}

// Int evaluates an integer flag.
func (c *client) Int(ctx context.Context, key string, defaultValue int) int {
	eval, err := c.evaluate(ctx, key, defaultValue)
	if err != nil {
		return defaultValue
	}

	switch v := eval.Value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}

	return defaultValue
}

// Float evaluates a float flag.
func (c *client) Float(ctx context.Context, key string, defaultValue float64) float64 {
	eval, err := c.evaluate(ctx, key, defaultValue)
	if err != nil {
		return defaultValue
	}

	switch v := eval.Value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}

	return defaultValue
}

// JSON evaluates a JSON flag into target.
func (c *client) JSON(ctx context.Context, key string, target interface{}) error {
	eval, err := c.evaluate(ctx, key, nil)
	if err != nil {
		return err
	}

	// Marshal and unmarshal to convert to target type
	data, err := json.Marshal(eval.Value)
	if err != nil {
		return fmt.Errorf("marshaling flag value: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshaling to target: %w", err)
	}

	return nil
}

// Variation evaluates a flag with full details.
func (c *client) Variation(ctx context.Context, key string) (*Evaluation, error) {
	return c.evaluate(ctx, key, nil)
}

// AllFlags returns all flag values for a context.
func (c *client) AllFlags(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})

	c.mu.RLock()
	provider := c.provider
	c.mu.RUnlock()

	if provider == nil {
		return result
	}

	flags, err := provider.GetAllFlags(ctx)
	if err != nil {
		return result
	}

	fctx := c.getContext(ctx)
	for key, flag := range flags {
		eval, err := c.evaluateFlag(ctx, flag, fctx)
		if err == nil {
			result[key] = eval.Value
		}
	}

	return result
}

// Track records a custom event for analytics.
func (c *client) Track(ctx context.Context, event string, data map[string]interface{}) {
	if c.options.eventHandler == nil {
		return
	}

	fctx := c.getContext(ctx)
	c.options.eventHandler(Event{
		Type:      "track",
		Key:       event,
		Context:   fctx,
		Timestamp: time.Now().Unix(),
		Data:      data,
	})
}

// Close releases resources.
func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.provider != nil {
		return c.provider.Close()
	}

	return nil
}

// evaluate performs flag evaluation.
func (c *client) evaluate(ctx context.Context, key string, defaultValue interface{}) (*Evaluation, error) {
	// Check for overrides first
	if override, ok := c.options.overrides[key]; ok {
		return &Evaluation{
			Key:          key,
			Value:        override,
			VariationIdx: -1,
			Reason:       ReasonTargetMatch,
		}, nil
	}

	c.mu.RLock()
	provider := c.provider
	c.mu.RUnlock()

	if provider == nil {
		return &Evaluation{
			Key:    key,
			Value:  defaultValue,
			Reason: ReasonError,
		}, ErrProviderNotInitialized
	}

	flag, err := provider.GetFlag(ctx, key)
	if err != nil {
		return &Evaluation{
			Key:    key,
			Value:  defaultValue,
			Reason: ReasonError,
		}, err
	}

	fctx := c.getContext(ctx)
	eval, err := c.evaluateFlag(ctx, flag, fctx)
	if err != nil {
		return &Evaluation{
			Key:    key,
			Value:  defaultValue,
			Reason: ReasonError,
		}, err
	}

	// Send evaluation event
	if c.options.eventHandler != nil {
		c.options.eventHandler(Event{
			Type:      "evaluation",
			Key:       key,
			Value:     eval.Value,
			Context:   fctx,
			Timestamp: time.Now().Unix(),
		})
	}

	return eval, nil
}

// evaluateFlag evaluates a single flag.
func (c *client) evaluateFlag(ctx context.Context, flag *Flag, fctx *Context) (*Evaluation, error) {
	if flag == nil {
		return nil, ErrFlagNotFound
	}

	// Check if flag is enabled
	if !flag.Enabled {
		return &Evaluation{
			Key:          flag.Key,
			Value:        flag.DefaultValue,
			VariationIdx: -1,
			Reason:       ReasonOff,
		}, nil
	}

	// Check prerequisites
	if len(flag.Prerequisites) > 0 {
		for _, prereq := range flag.Prerequisites {
			prereqEval, err := c.evaluate(ctx, prereq.Key, nil)
			if err != nil {
				return &Evaluation{
					Key:                flag.Key,
					Value:              flag.DefaultValue,
					VariationIdx:       -1,
					Reason:             ReasonPrerequisite,
					PrerequisiteFailed: true,
				}, nil
			}

			if prereqEval.VariationIdx != prereq.Variation {
				return &Evaluation{
					Key:                flag.Key,
					Value:              flag.DefaultValue,
					VariationIdx:       -1,
					Reason:             ReasonPrerequisite,
					PrerequisiteFailed: true,
				}, nil
			}
		}
	}

	// Evaluate rules
	for _, rule := range flag.Rules {
		if rule.evaluate(fctx) {
			// Rule matched
			if rule.Rollout != nil {
				// Percentage rollout
				variationIdx := c.evaluateRollout(rule.Rollout, fctx, flag.Key)
				return &Evaluation{
					Key:          flag.Key,
					Value:        flag.getVariantValue(variationIdx),
					VariationIdx: variationIdx,
					Reason:       ReasonRuleMatch,
					RuleID:       rule.ID,
					InExperiment: true,
				}, nil
			}

			// Check percentage
			if rule.Percentage != nil {
				bucket := c.getBucket(fctx, flag.Key)
				if bucket >= *rule.Percentage {
					continue // Skip this rule
				}
			}

			return &Evaluation{
				Key:          flag.Key,
				Value:        flag.getVariantValue(rule.Variation),
				VariationIdx: rule.Variation,
				Reason:       ReasonRuleMatch,
				RuleID:       rule.ID,
			}, nil
		}
	}

	// Fallthrough to default
	return &Evaluation{
		Key:          flag.Key,
		Value:        flag.DefaultValue,
		VariationIdx: -1,
		Reason:       ReasonFallthrough,
	}, nil
}

// evaluateRollout determines which variation based on percentage weights.
func (c *client) evaluateRollout(rollout *Rollout, fctx *Context, flagKey string) int {
	if rollout == nil || len(rollout.Variations) == 0 {
		return 0
	}

	// Get consistent bucket value (0-100000)
	bucket := c.getBucketValue(fctx, rollout.BucketBy, flagKey)

	// Find variation based on weight
	cumulative := 0
	for _, wv := range rollout.Variations {
		cumulative += wv.Weight
		if bucket < cumulative {
			return wv.Variation
		}
	}

	// Default to first variation
	return rollout.Variations[0].Variation
}

// getBucket returns a bucket value 0-99 for percentage checks.
func (c *client) getBucket(fctx *Context, flagKey string) int {
	return c.getBucketValue(fctx, "key", flagKey) / 1000
}

// getBucketValue returns a consistent bucket value 0-99999 based on context attribute.
func (c *client) getBucketValue(fctx *Context, bucketBy string, flagKey string) int {
	if fctx == nil {
		return 0
	}

	var value string
	if bucketBy == "" || bucketBy == "key" {
		value = fctx.Key
	} else {
		attr := fctx.getAttribute(bucketBy)
		if attr != nil {
			value = fmt.Sprintf("%v", attr)
		}
	}

	if value == "" {
		return 0
	}

	// Create consistent hash
	h := fnv.New32a()
	h.Write([]byte(flagKey + ":" + value))
	return int(h.Sum32() % 100000)
}

// getContext extracts the feature context from the standard context.
func (c *client) getContext(ctx context.Context) *Context {
	if fctx := FromContext(ctx); fctx != nil {
		return fctx
	}

	if c.options.defaultContext != nil {
		return c.options.defaultContext
	}

	return NewContext("")
}

// SetProvider sets the provider for the client.
func (c *client) SetProvider(provider Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.provider = provider
}
