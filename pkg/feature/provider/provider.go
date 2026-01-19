package provider

import (
	"context"

	"github.com/rompi/core-backend/pkg/feature"
)

// Provider defines the interface for feature flag backends.
type Provider interface {
	// GetFlag retrieves a flag by key.
	GetFlag(ctx context.Context, key string) (*feature.Flag, error)

	// GetAllFlags retrieves all flags.
	GetAllFlags(ctx context.Context) (map[string]*feature.Flag, error)

	// SetFlag creates or updates a flag.
	SetFlag(ctx context.Context, flag *feature.Flag) error

	// DeleteFlag removes a flag.
	DeleteFlag(ctx context.Context, key string) error

	// Close releases provider resources.
	Close() error
}

// Refreshable is implemented by providers that support live updates.
type Refreshable interface {
	// Refresh reloads flags from the backend.
	Refresh(ctx context.Context) error

	// OnUpdate registers a callback for flag updates.
	OnUpdate(callback func(flags map[string]*feature.Flag))
}

// Watchable is implemented by providers that support watching for changes.
type Watchable interface {
	// Watch starts watching for flag changes.
	Watch(ctx context.Context) error

	// StopWatching stops watching for changes.
	StopWatching()
}
