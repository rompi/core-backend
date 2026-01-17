package provider

import (
	"context"
	"sync"
	"testing"

	"github.com/rompi/core-backend/pkg/feature"
)

func TestMemoryProvider_NewMemoryProvider(t *testing.T) {
	p := NewMemoryProvider()
	if p == nil {
		t.Fatal("NewMemoryProvider() returned nil")
	}
	if p.flags == nil {
		t.Error("flags map should be initialized")
	}
}

func TestMemoryProvider_NewMemoryProviderWithFlags(t *testing.T) {
	flags := map[string]*feature.Flag{
		"flag-1": {Key: "flag-1", Enabled: true},
		"flag-2": {Key: "flag-2", Enabled: false},
	}

	p := NewMemoryProviderWithFlags(flags)

	allFlags, err := p.GetAllFlags(context.Background())
	if err != nil {
		t.Fatalf("GetAllFlags() error = %v", err)
	}
	if len(allFlags) != 2 {
		t.Errorf("GetAllFlags() returned %d flags, want 2", len(allFlags))
	}
}

func TestMemoryProvider_GetFlag(t *testing.T) {
	p := NewMemoryProvider()
	testFlag := &feature.Flag{Key: "test-flag", Enabled: true}
	_ = p.SetFlag(context.Background(), testFlag)

	// Test getting existing flag
	flag, err := p.GetFlag(context.Background(), "test-flag")
	if err != nil {
		t.Errorf("GetFlag() error = %v", err)
	}
	if flag.Key != "test-flag" {
		t.Errorf("GetFlag() key = %v, want test-flag", flag.Key)
	}

	// Test getting non-existent flag
	_, err = p.GetFlag(context.Background(), "non-existent")
	if err != feature.ErrFlagNotFound {
		t.Errorf("GetFlag() for non-existent = %v, want ErrFlagNotFound", err)
	}
}

func TestMemoryProvider_GetAllFlags(t *testing.T) {
	p := NewMemoryProvider()
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "flag-1"})
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "flag-2"})
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "flag-3"})

	flags, err := p.GetAllFlags(context.Background())
	if err != nil {
		t.Errorf("GetAllFlags() error = %v", err)
	}
	if len(flags) != 3 {
		t.Errorf("GetAllFlags() returned %d flags, want 3", len(flags))
	}
}

func TestMemoryProvider_SetFlag(t *testing.T) {
	p := NewMemoryProvider()

	// Test setting valid flag
	err := p.SetFlag(context.Background(), &feature.Flag{Key: "new-flag", Enabled: true})
	if err != nil {
		t.Errorf("SetFlag() error = %v", err)
	}

	flag, _ := p.GetFlag(context.Background(), "new-flag")
	if !flag.Enabled {
		t.Error("SetFlag() did not save enabled state")
	}

	// Test setting nil flag
	err = p.SetFlag(context.Background(), nil)
	if err != feature.ErrInvalidConfig {
		t.Errorf("SetFlag(nil) error = %v, want ErrInvalidConfig", err)
	}

	// Test setting flag with empty key
	err = p.SetFlag(context.Background(), &feature.Flag{Key: ""})
	if err != feature.ErrInvalidConfig {
		t.Errorf("SetFlag(empty key) error = %v, want ErrInvalidConfig", err)
	}
}

func TestMemoryProvider_DeleteFlag(t *testing.T) {
	p := NewMemoryProvider()
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "to-delete"})

	// Verify flag exists
	_, err := p.GetFlag(context.Background(), "to-delete")
	if err != nil {
		t.Fatal("Flag should exist before delete")
	}

	// Delete flag
	err = p.DeleteFlag(context.Background(), "to-delete")
	if err != nil {
		t.Errorf("DeleteFlag() error = %v", err)
	}

	// Verify flag is gone
	_, err = p.GetFlag(context.Background(), "to-delete")
	if err != feature.ErrFlagNotFound {
		t.Errorf("GetFlag() after delete = %v, want ErrFlagNotFound", err)
	}
}

func TestMemoryProvider_Close(t *testing.T) {
	p := NewMemoryProvider()
	err := p.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestMemoryProvider_OnUpdate(t *testing.T) {
	p := NewMemoryProvider()

	var callbackInvoked bool
	var receivedFlags map[string]*feature.Flag

	p.OnUpdate(func(flags map[string]*feature.Flag) {
		callbackInvoked = true
		receivedFlags = flags
	})

	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "callback-test"})

	if !callbackInvoked {
		t.Error("OnUpdate callback should have been invoked")
	}
	if receivedFlags == nil {
		t.Error("Callback should have received flags")
	}
	if _, ok := receivedFlags["callback-test"]; !ok {
		t.Error("Callback should have received the new flag")
	}
}

func TestMemoryProvider_LoadFlags(t *testing.T) {
	p := NewMemoryProvider()

	flags := []*feature.Flag{
		{Key: "flag-1"},
		{Key: "flag-2"},
		{Key: "flag-3"},
		nil, // Should be skipped
		{Key: ""}, // Should be skipped
	}

	p.LoadFlags(flags)

	allFlags, _ := p.GetAllFlags(context.Background())
	if len(allFlags) != 3 {
		t.Errorf("LoadFlags() resulted in %d flags, want 3", len(allFlags))
	}
}

func TestMemoryProvider_Clear(t *testing.T) {
	p := NewMemoryProvider()
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "flag-1"})
	_ = p.SetFlag(context.Background(), &feature.Flag{Key: "flag-2"})

	p.Clear()

	flags, _ := p.GetAllFlags(context.Background())
	if len(flags) != 0 {
		t.Errorf("Clear() should remove all flags, got %d", len(flags))
	}
}

func TestMemoryProvider_ConcurrentAccess(t *testing.T) {
	p := NewMemoryProvider()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + (i % 26)))
			_ = p.SetFlag(ctx, &feature.Flag{Key: key})
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = p.GetAllFlags(ctx)
		}()
	}
	wg.Wait()

	// Concurrent read/write mix
	wg.Add(numGoroutines * 2)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			_ = p.SetFlag(ctx, &feature.Flag{Key: "concurrent"})
		}(i)
		go func() {
			defer wg.Done()
			_, _ = p.GetFlag(ctx, "concurrent")
		}()
	}
	wg.Wait()
}
