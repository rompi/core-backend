package provider

import (
	"context"
	"sync"
)

// MemoryProvider stores configuration in memory.
// This is primarily useful for testing.
type MemoryProvider struct {
	mu     sync.RWMutex
	values map[string]any
}

// NewMemoryProvider creates a new in-memory provider.
func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{
		values: make(map[string]any),
	}
}

// NewMemoryProviderWithValues creates a new in-memory provider with initial values.
func NewMemoryProviderWithValues(values map[string]any) *MemoryProvider {
	p := &MemoryProvider{
		values: make(map[string]any),
	}

	for k, v := range values {
		p.values[k] = v
	}

	return p
}

// Name returns the provider name.
func (p *MemoryProvider) Name() string {
	return "memory"
}

// Load returns the stored configuration.
func (p *MemoryProvider) Load(_ context.Context) (map[string]any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]any, len(p.values))
	for k, v := range p.values {
		result[k] = v
	}

	return result, nil
}

// Watch is not implemented for memory provider.
func (p *MemoryProvider) Watch(_ context.Context, _ func()) error {
	return nil
}

// Set sets a value in the memory provider.
func (p *MemoryProvider) Set(key string, value any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.values[key] = value
}

// Delete removes a value from the memory provider.
func (p *MemoryProvider) Delete(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.values, key)
}

// Clear removes all values from the memory provider.
func (p *MemoryProvider) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.values = make(map[string]any)
}
