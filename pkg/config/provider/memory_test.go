package provider

import (
	"context"
	"testing"
)

func TestNewMemoryProvider(t *testing.T) {
	p := NewMemoryProvider()
	if p == nil {
		t.Fatal("NewMemoryProvider() returned nil")
	}

	if p.Name() != "memory" {
		t.Errorf("Name() = %v, want %v", p.Name(), "memory")
	}
}

func TestNewMemoryProviderWithValues(t *testing.T) {
	values := map[string]any{
		"host": "localhost",
		"port": 8080,
	}

	p := NewMemoryProviderWithValues(values)
	if p == nil {
		t.Fatal("NewMemoryProviderWithValues() returned nil")
	}

	loaded, err := p.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded["host"] != "localhost" {
		t.Errorf("loaded[host] = %v, want %v", loaded["host"], "localhost")
	}
	if loaded["port"] != 8080 {
		t.Errorf("loaded[port] = %v, want %v", loaded["port"], 8080)
	}
}

func TestMemoryProvider_Load(t *testing.T) {
	p := NewMemoryProvider()
	p.Set("key", "value")

	loaded, err := p.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded["key"] != "value" {
		t.Errorf("loaded[key] = %v, want %v", loaded["key"], "value")
	}
}

func TestMemoryProvider_Load_ReturnsACopy(t *testing.T) {
	p := NewMemoryProvider()
	p.Set("key", "value")

	loaded, _ := p.Load(context.Background())
	loaded["key"] = "modified"

	// Original should not be modified
	loaded2, _ := p.Load(context.Background())
	if loaded2["key"] != "value" {
		t.Errorf("Load() should return a copy, original was modified")
	}
}

func TestMemoryProvider_Set(t *testing.T) {
	p := NewMemoryProvider()
	p.Set("key", "value")

	loaded, _ := p.Load(context.Background())
	if loaded["key"] != "value" {
		t.Errorf("loaded[key] = %v, want %v", loaded["key"], "value")
	}
}

func TestMemoryProvider_Delete(t *testing.T) {
	p := NewMemoryProvider()
	p.Set("key", "value")
	p.Delete("key")

	loaded, _ := p.Load(context.Background())
	if _, exists := loaded["key"]; exists {
		t.Error("Delete() did not remove the key")
	}
}

func TestMemoryProvider_Clear(t *testing.T) {
	p := NewMemoryProvider()
	p.Set("key1", "value1")
	p.Set("key2", "value2")
	p.Clear()

	loaded, _ := p.Load(context.Background())
	if len(loaded) != 0 {
		t.Errorf("Clear() did not remove all keys, len = %d", len(loaded))
	}
}

func TestMemoryProvider_Watch(t *testing.T) {
	p := NewMemoryProvider()

	// Watch should return nil (not supported)
	err := p.Watch(context.Background(), func() {})
	if err != nil {
		t.Errorf("Watch() error = %v, want nil", err)
	}
}
