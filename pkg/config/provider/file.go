package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Parser parses configuration data from bytes.
type Parser interface {
	// Parse parses data into a configuration map.
	Parse(data []byte) (map[string]any, error)

	// Extensions returns the file extensions this parser supports.
	Extensions() []string
}

// FileOption configures the FileProvider.
type FileOption func(*FileProvider)

// FileProvider reads configuration from a file.
type FileProvider struct {
	path          string
	parser        Parser
	optional      bool
	watchInterval time.Duration

	mu          sync.RWMutex
	lastModTime time.Time
}

// NewFileProvider creates a new file-based configuration provider.
// The file format is auto-detected from the file extension unless
// a specific parser is provided via WithParser.
func NewFileProvider(path string, opts ...FileOption) *FileProvider {
	p := &FileProvider{
		path:          path,
		watchInterval: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithParser sets a specific parser for the file.
func WithParser(parser Parser) FileOption {
	return func(p *FileProvider) {
		p.parser = parser
	}
}

// WithOptional marks the file as optional.
// If the file doesn't exist, Load will return an empty map instead of an error.
func WithOptional() FileOption {
	return func(p *FileProvider) {
		p.optional = true
	}
}

// WithFileWatchInterval sets the interval for watching file changes.
func WithFileWatchInterval(interval time.Duration) FileOption {
	return func(p *FileProvider) {
		if interval > 0 {
			p.watchInterval = interval
		}
	}
}

// Name returns the provider name.
func (p *FileProvider) Name() string {
	return "file:" + filepath.Base(p.path)
}

// Load reads and parses the configuration file.
func (p *FileProvider) Load(_ context.Context) (map[string]any, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) && p.optional {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("failed to read file %s: %w", p.path, err)
	}

	// Update last modification time
	if info, err := os.Stat(p.path); err == nil {
		p.mu.Lock()
		p.lastModTime = info.ModTime()
		p.mu.Unlock()
	}

	parser := p.getParser()
	if parser == nil {
		return nil, fmt.Errorf("no parser available for file %s", p.path)
	}

	result, err := parser.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", p.path, err)
	}

	return result, nil
}

// Watch watches for file changes and calls the callback when changes are detected.
func (p *FileProvider) Watch(ctx context.Context, callback func()) error {
	go func() {
		ticker := time.NewTicker(p.watchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if p.hasChanged() {
					callback()
				}
			}
		}
	}()

	return nil
}

// hasChanged checks if the file has been modified since last read.
func (p *FileProvider) hasChanged() bool {
	info, err := os.Stat(p.path)
	if err != nil {
		return false
	}

	p.mu.RLock()
	lastMod := p.lastModTime
	p.mu.RUnlock()

	if info.ModTime().After(lastMod) {
		p.mu.Lock()
		p.lastModTime = info.ModTime()
		p.mu.Unlock()
		return true
	}

	return false
}

// getParser returns the parser for the file.
func (p *FileProvider) getParser() Parser {
	if p.parser != nil {
		return p.parser
	}

	// Auto-detect from extension
	ext := strings.ToLower(filepath.Ext(p.path))

	switch ext {
	case ".json":
		return &jsonParser{}
	case ".yaml", ".yml":
		return &yamlParser{}
	default:
		return nil
	}
}

// jsonParser is a simple JSON parser for auto-detection.
type jsonParser struct{}

func (p *jsonParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := jsonUnmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *jsonParser) Extensions() []string {
	return []string{".json"}
}

// yamlParser is a simple YAML parser for auto-detection.
type yamlParser struct{}

func (p *yamlParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yamlUnmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *yamlParser) Extensions() []string {
	return []string{".yaml", ".yml"}
}
