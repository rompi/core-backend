package provider

import (
	"context"
	"sync"

	"github.com/rompi/core-backend/pkg/feature"
)

// MemoryProvider is an in-memory feature flag provider.
// It is safe for concurrent use.
type MemoryProvider struct {
	mu       sync.RWMutex
	flags    map[string]*feature.Flag
	callback func(flags map[string]*feature.Flag)
}

// NewMemoryProvider creates a new in-memory provider.
func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{
		flags: make(map[string]*feature.Flag),
	}
}

// NewMemoryProviderWithFlags creates a new in-memory provider with initial flags.
func NewMemoryProviderWithFlags(flags map[string]*feature.Flag) *MemoryProvider {
	p := NewMemoryProvider()
	for k, v := range flags {
		p.flags[k] = v
	}
	return p
}

// GetFlag retrieves a flag by key.
func (p *MemoryProvider) GetFlag(ctx context.Context, key string) (*feature.Flag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	flag, ok := p.flags[key]
	if !ok {
		return nil, feature.ErrFlagNotFound
	}

	return flag, nil
}

// GetAllFlags retrieves all flags.
func (p *MemoryProvider) GetAllFlags(ctx context.Context) (map[string]*feature.Flag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*feature.Flag, len(p.flags))
	for k, v := range p.flags {
		result[k] = v
	}

	return result, nil
}

// SetFlag creates or updates a flag.
func (p *MemoryProvider) SetFlag(ctx context.Context, flag *feature.Flag) error {
	if flag == nil || flag.Key == "" {
		return feature.ErrInvalidConfig
	}

	p.mu.Lock()
	p.flags[flag.Key] = flag
	p.mu.Unlock()

	p.notifyUpdate()
	return nil
}

// DeleteFlag removes a flag.
func (p *MemoryProvider) DeleteFlag(ctx context.Context, key string) error {
	p.mu.Lock()
	delete(p.flags, key)
	p.mu.Unlock()

	p.notifyUpdate()
	return nil
}

// Close releases provider resources.
func (p *MemoryProvider) Close() error {
	return nil
}

// OnUpdate registers a callback for flag updates.
func (p *MemoryProvider) OnUpdate(callback func(flags map[string]*feature.Flag)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callback = callback
}

// notifyUpdate calls the update callback if registered.
func (p *MemoryProvider) notifyUpdate() {
	p.mu.RLock()
	callback := p.callback
	p.mu.RUnlock()

	if callback != nil {
		flags, _ := p.GetAllFlags(context.Background())
		callback(flags)
	}
}

// LoadFlags loads multiple flags at once.
func (p *MemoryProvider) LoadFlags(flags []*feature.Flag) {
	p.mu.Lock()
	for _, flag := range flags {
		if flag != nil && flag.Key != "" {
			p.flags[flag.Key] = flag
		}
	}
	p.mu.Unlock()

	p.notifyUpdate()
}

// Clear removes all flags.
func (p *MemoryProvider) Clear() {
	p.mu.Lock()
	p.flags = make(map[string]*feature.Flag)
	p.mu.Unlock()

	p.notifyUpdate()
}
