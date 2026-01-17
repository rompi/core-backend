package config

import (
	"context"
	"sync"
	"time"
)

// Watcher watches for configuration changes from providers.
type Watcher struct {
	config        *configImpl
	interval      time.Duration
	callbacks     []func(Config)
	mu            sync.RWMutex
	stopCh        chan struct{}
	running       bool
}

// newWatcher creates a new watcher instance.
func newWatcher(config *configImpl, interval time.Duration) *Watcher {
	return &Watcher{
		config:    config,
		interval:  interval,
		callbacks: []func(Config){},
		stopCh:    make(chan struct{}),
	}
}

// OnChange registers a callback to be called when configuration changes.
func (w *Watcher) OnChange(callback func(Config)) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.callbacks = append(w.callbacks, callback)
}

// Start starts watching for configuration changes.
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	// Start watching each provider
	for _, provider := range w.config.providers {
		providerCallback := func() {
			w.notifyCallbacks()
		}

		// Start provider-specific watching
		if err := provider.Watch(ctx, providerCallback); err != nil {
			// Not all providers support watching, continue
			continue
		}
	}

	return nil
}

// Stop stops watching for configuration changes.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopCh)
	w.running = false
}

// notifyCallbacks notifies all registered callbacks.
func (w *Watcher) notifyCallbacks() {
	w.mu.RLock()
	callbacks := make([]func(Config), len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	// Reload configuration
	ctx := context.Background()
	if err := w.config.Load(ctx); err != nil {
		return
	}

	// Notify callbacks
	for _, callback := range callbacks {
		callback(w.config)
	}
}

// WatchConfig is a helper function to start watching configuration changes.
// It returns a cancel function that stops watching.
func WatchConfig(ctx context.Context, cfg Config, callback func(Config)) (func(), error) {
	impl, ok := cfg.(*configImpl)
	if !ok {
		// For sub-configs or other implementations, delegate to Watch method
		return func() {}, cfg.Watch(ctx, callback)
	}

	watcher := newWatcher(impl, impl.watchInterval)
	watcher.OnChange(callback)

	if err := watcher.Start(ctx); err != nil {
		return nil, err
	}

	cancel := func() {
		watcher.Stop()
	}

	return cancel, nil
}
