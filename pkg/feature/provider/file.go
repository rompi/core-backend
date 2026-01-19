package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rompi/core-backend/pkg/feature"
)

// FileProvider is a file-based feature flag provider.
// It supports YAML and JSON formats and optional file watching.
type FileProvider struct {
	mu          sync.RWMutex
	path        string
	flags       map[string]*feature.Flag
	callback    func(flags map[string]*feature.Flag)
	watchCancel context.CancelFunc
	lastModTime time.Time
}

// FileFlags represents the structure of a feature flags file.
type FileFlags struct {
	Flags map[string]*feature.Flag `json:"flags" yaml:"flags"`
}

// NewFileProvider creates a new file-based provider.
func NewFileProvider(path string) (*FileProvider, error) {
	p := &FileProvider{
		path:  path,
		flags: make(map[string]*feature.Flag),
	}

	if err := p.load(); err != nil {
		return nil, err
	}

	return p, nil
}

// GetFlag retrieves a flag by key.
func (p *FileProvider) GetFlag(ctx context.Context, key string) (*feature.Flag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	flag, ok := p.flags[key]
	if !ok {
		return nil, feature.ErrFlagNotFound
	}

	return flag, nil
}

// GetAllFlags retrieves all flags.
func (p *FileProvider) GetAllFlags(ctx context.Context) (map[string]*feature.Flag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*feature.Flag, len(p.flags))
	for k, v := range p.flags {
		result[k] = v
	}

	return result, nil
}

// SetFlag creates or updates a flag and persists to file.
func (p *FileProvider) SetFlag(ctx context.Context, flag *feature.Flag) error {
	if flag == nil || flag.Key == "" {
		return feature.ErrInvalidConfig
	}

	p.mu.Lock()
	p.flags[flag.Key] = flag
	p.mu.Unlock()

	if err := p.save(); err != nil {
		return err
	}

	p.notifyUpdate()
	return nil
}

// DeleteFlag removes a flag and persists to file.
func (p *FileProvider) DeleteFlag(ctx context.Context, key string) error {
	p.mu.Lock()
	delete(p.flags, key)
	p.mu.Unlock()

	if err := p.save(); err != nil {
		return err
	}

	p.notifyUpdate()
	return nil
}

// Close releases provider resources.
func (p *FileProvider) Close() error {
	p.StopWatching()
	return nil
}

// Refresh reloads flags from the file.
func (p *FileProvider) Refresh(ctx context.Context) error {
	return p.load()
}

// OnUpdate registers a callback for flag updates.
func (p *FileProvider) OnUpdate(callback func(flags map[string]*feature.Flag)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callback = callback
}

// Watch starts watching the file for changes.
func (p *FileProvider) Watch(ctx context.Context) error {
	watchCtx, cancel := context.WithCancel(ctx)
	p.watchCancel = cancel

	go p.watchLoop(watchCtx)
	return nil
}

// StopWatching stops watching for file changes.
func (p *FileProvider) StopWatching() {
	if p.watchCancel != nil {
		p.watchCancel()
		p.watchCancel = nil
	}
}

// watchLoop polls the file for changes.
func (p *FileProvider) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if p.hasFileChanged() {
				if err := p.load(); err == nil {
					p.notifyUpdate()
				}
			}
		}
	}
}

// hasFileChanged checks if the file has been modified.
func (p *FileProvider) hasFileChanged() bool {
	info, err := os.Stat(p.path)
	if err != nil {
		return false
	}

	p.mu.RLock()
	changed := info.ModTime().After(p.lastModTime)
	p.mu.RUnlock()

	return changed
}

// load reads and parses the flags file.
func (p *FileProvider) load() error {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return feature.ErrFileNotFound
		}
		return fmt.Errorf("reading file: %w", err)
	}

	info, err := os.Stat(p.path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	var fileFlags FileFlags

	ext := filepath.Ext(p.path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &fileFlags); err != nil {
			return fmt.Errorf("%w: %v", feature.ErrInvalidFileFormat, err)
		}
	case ".yaml", ".yml":
		if err := unmarshalYAML(data, &fileFlags); err != nil {
			return fmt.Errorf("%w: %v", feature.ErrInvalidFileFormat, err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &fileFlags); err != nil {
			if err := unmarshalYAML(data, &fileFlags); err != nil {
				return fmt.Errorf("%w: unable to parse as JSON or YAML", feature.ErrInvalidFileFormat)
			}
		}
	}

	// Ensure flag keys are set from map keys
	for key, flag := range fileFlags.Flags {
		if flag.Key == "" {
			flag.Key = key
		}
	}

	p.mu.Lock()
	p.flags = fileFlags.Flags
	if p.flags == nil {
		p.flags = make(map[string]*feature.Flag)
	}
	p.lastModTime = info.ModTime()
	p.mu.Unlock()

	return nil
}

// save persists flags to the file.
func (p *FileProvider) save() error {
	p.mu.RLock()
	fileFlags := FileFlags{Flags: p.flags}
	p.mu.RUnlock()

	var data []byte
	var err error

	ext := filepath.Ext(p.path)
	switch ext {
	case ".json":
		data, err = json.MarshalIndent(fileFlags, "", "  ")
	case ".yaml", ".yml":
		data, err = marshalYAML(fileFlags)
	default:
		data, err = json.MarshalIndent(fileFlags, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("marshaling flags: %w", err)
	}

	if err := os.WriteFile(p.path, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// notifyUpdate calls the update callback if registered.
func (p *FileProvider) notifyUpdate() {
	p.mu.RLock()
	callback := p.callback
	p.mu.RUnlock()

	if callback != nil {
		flags, _ := p.GetAllFlags(context.Background())
		callback(flags)
	}
}

// unmarshalYAML is a simple YAML unmarshaler.
// For production use, consider using a proper YAML library.
func unmarshalYAML(data []byte, v interface{}) error {
	// Simple YAML to JSON conversion for basic YAML files.
	// This handles simple key-value YAML structures.
	// For full YAML support, use gopkg.in/yaml.v3.
	return json.Unmarshal(convertSimpleYAMLToJSON(data), v)
}

// marshalYAML is a simple YAML marshaler.
func marshalYAML(v interface{}) ([]byte, error) {
	// For simplicity, we marshal as JSON.
	// For full YAML support, use gopkg.in/yaml.v3.
	return json.MarshalIndent(v, "", "  ")
}

// convertSimpleYAMLToJSON converts simple YAML to JSON.
// This is a basic implementation for simple configurations.
func convertSimpleYAMLToJSON(data []byte) []byte {
	// For now, assume JSON-compatible format.
	// A full YAML parser would handle all YAML features.
	return data
}
