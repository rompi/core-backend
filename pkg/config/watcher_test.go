package config

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	impl := &configImpl{
		providers:     []Provider{},
		values:        make(map[string]any),
		watchInterval: 30 * time.Second,
	}

	watcher := newWatcher(impl, 10*time.Second)

	if watcher == nil {
		t.Fatal("newWatcher() returned nil")
	}

	if watcher.interval != 10*time.Second {
		t.Errorf("watcher.interval = %v, want %v", watcher.interval, 10*time.Second)
	}
}

func TestWatcher_OnChange(t *testing.T) {
	impl := &configImpl{
		providers:     []Provider{},
		values:        make(map[string]any),
		watchInterval: 30 * time.Second,
	}

	watcher := newWatcher(impl, 10*time.Second)

	callCount := 0
	watcher.OnChange(func(cfg Config) {
		callCount++
	})

	if len(watcher.callbacks) != 1 {
		t.Errorf("OnChange() did not register callback, len = %d", len(watcher.callbacks))
	}
}

func TestWatcher_StartStop(t *testing.T) {
	impl := &configImpl{
		providers:     []Provider{},
		values:        make(map[string]any),
		watchInterval: 30 * time.Second,
		keyDelimiter:  ".",
	}

	watcher := newWatcher(impl, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := watcher.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !watcher.running {
		t.Error("Start() did not set running to true")
	}

	// Start again should be a no-op
	err = watcher.Start(ctx)
	if err != nil {
		t.Fatalf("Start() again error = %v", err)
	}

	watcher.Stop()

	if watcher.running {
		t.Error("Stop() did not set running to false")
	}

	// Stop again should be a no-op
	watcher.Stop()
}

func TestWatcher_NotifyCallbacks(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"key": "value"},
	}

	impl := &configImpl{
		providers:     []Provider{provider},
		values:        make(map[string]any),
		watchInterval: 30 * time.Second,
		keyDelimiter:  ".",
	}

	watcher := newWatcher(impl, 10*time.Millisecond)

	var callCount int32
	watcher.OnChange(func(cfg Config) {
		atomic.AddInt32(&callCount, 1)
	})

	// Manually trigger notify
	watcher.notifyCallbacks()

	time.Sleep(10 * time.Millisecond) // Give goroutine time to complete

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("notifyCallbacks() did not call callback, count = %d", callCount)
	}
}

func TestWatchConfig(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"key": "value"},
	}

	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopFunc, err := WatchConfig(ctx, cfg, func(c Config) {
		// Callback
	})

	if err != nil {
		t.Fatalf("WatchConfig() error = %v", err)
	}

	if stopFunc == nil {
		t.Error("WatchConfig() returned nil stop function")
	}

	stopFunc()
}

func TestWatchConfig_SubConfig(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"app.key": "value"},
	}

	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	subCfg := cfg.Sub("app")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopFunc, err := WatchConfig(ctx, subCfg, func(c Config) {
		// Callback
	})

	if err != nil {
		t.Fatalf("WatchConfig() error = %v", err)
	}

	if stopFunc == nil {
		t.Error("WatchConfig() returned nil stop function")
	}
}
