package catalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// EmbedCatalog implements Catalog using an embedded filesystem.
type EmbedCatalog struct {
	fsys     fs.FS
	root     string
	format   string // "json" or "yaml"
	messages map[string]map[string]*Message
	locales  []string
	mu       sync.RWMutex
}

// EmbedOption configures an EmbedCatalog.
type EmbedOption func(*EmbedCatalog)

// WithFormat sets the file format for the embedded catalog.
// Supported formats: "json" (default), "yaml".
func WithFormat(format string) EmbedOption {
	return func(c *EmbedCatalog) {
		c.format = strings.ToLower(format)
	}
}

// NewEmbedCatalog creates a new catalog from an embedded filesystem.
// The fsys should contain translation files in the root directory.
// Example:
//
//	//go:embed locales/*.json
//	var localesFS embed.FS
//	cat, err := catalog.NewEmbedCatalog(localesFS, "locales")
func NewEmbedCatalog(fsys fs.FS, root string, opts ...EmbedOption) (*EmbedCatalog, error) {
	c := &EmbedCatalog{
		fsys:     fsys,
		root:     root,
		format:   "json",
		messages: make(map[string]map[string]*Message),
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.load(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFailed, err)
	}

	return c, nil
}

// Lookup finds a message by locale and key.
func (c *EmbedCatalog) Lookup(locale, key string) (*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	localeMessages, ok := c.messages[locale]
	if !ok {
		return nil, ErrLocaleNotFound
	}

	msg, ok := localeMessages[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return msg, nil
}

// All returns all messages for a locale.
func (c *EmbedCatalog) All(locale string) (map[string]*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	localeMessages, ok := c.messages[locale]
	if !ok {
		return nil, ErrLocaleNotFound
	}

	result := make(map[string]*Message, len(localeMessages))
	for k, v := range localeMessages {
		result[k] = v
	}

	return result, nil
}

// Locales returns all available locales.
func (c *EmbedCatalog) Locales() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.locales))
	copy(result, c.locales)
	return result
}

// Reload reloads the catalog from the embedded filesystem.
// Note: For embedded filesystems, this typically returns the same data.
func (c *EmbedCatalog) Reload() error {
	return c.load()
}

// load loads all translation files from the embedded filesystem.
func (c *EmbedCatalog) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newMessages := make(map[string]map[string]*Message)
	var newLocales []string

	ext := "." + c.format
	if c.format == "yaml" {
		ext = ".yaml" // Also check .yml below
	}

	err := fs.WalkDir(c.fsys, c.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		name := d.Name()
		isValidExt := strings.HasSuffix(name, ext)
		if c.format == "yaml" {
			isValidExt = isValidExt || strings.HasSuffix(name, ".yml")
		}

		if !isValidExt {
			return nil
		}

		// Extract locale from filename
		locale := strings.TrimSuffix(strings.TrimSuffix(name, ".json"), ".yaml")
		locale = strings.TrimSuffix(locale, ".yml")

		messages, err := c.loadFile(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		newMessages[locale] = messages
		newLocales = append(newLocales, locale)

		return nil
	})

	if err != nil {
		return err
	}

	c.messages = newMessages
	c.locales = newLocales

	return nil
}

// loadFile loads a single file from the embedded filesystem.
func (c *EmbedCatalog) loadFile(path string) (map[string]*Message, error) {
	data, err := fs.ReadFile(c.fsys, path)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}

	switch c.format {
	case "json":
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case "yaml":
		// Import yaml dynamically would require build tag, so we support JSON by default
		// For YAML in embedded files, users should use JSON format or provide custom catalog
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", c.format)
	}

	messages := make(map[string]*Message)
	c.parseMessages("", raw, messages)

	return messages, nil
}

// parseMessages recursively parses data into messages.
func (c *EmbedCatalog) parseMessages(prefix string, data map[string]interface{}, messages map[string]*Message) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			messages[fullKey] = &Message{
				ID:    fullKey,
				Other: v,
			}

		case map[string]interface{}:
			if c.isPluralObject(v) {
				msg := c.parsePluralMessage(fullKey, v)
				messages[fullKey] = msg
			} else {
				c.parseMessages(fullKey, v, messages)
			}
		}
	}
}

// isPluralObject checks if the map contains plural form keys.
func (c *EmbedCatalog) isPluralObject(m map[string]interface{}) bool {
	pluralKeys := []string{"one", "other", "zero", "two", "few", "many"}
	for _, pk := range pluralKeys {
		if _, ok := m[pk]; ok {
			return true
		}
	}
	return false
}

// parsePluralMessage creates a Message from a plural object.
func (c *EmbedCatalog) parsePluralMessage(key string, m map[string]interface{}) *Message {
	msg := &Message{ID: key}

	if v, ok := m["one"].(string); ok {
		msg.One = v
	}
	if v, ok := m["other"].(string); ok {
		msg.Other = v
	}
	if v, ok := m["zero"].(string); ok {
		msg.Zero = v
	}
	if v, ok := m["two"].(string); ok {
		msg.Two = v
	}
	if v, ok := m["few"].(string); ok {
		msg.Few = v
	}
	if v, ok := m["many"].(string); ok {
		msg.Many = v
	}
	if v, ok := m["description"].(string); ok {
		msg.Description = v
	}

	return msg
}

// InMemoryCatalog implements Catalog with in-memory storage.
// Useful for testing or dynamic translations.
type InMemoryCatalog struct {
	messages map[string]map[string]*Message
	mu       sync.RWMutex
}

// NewInMemoryCatalog creates a new in-memory catalog.
func NewInMemoryCatalog() *InMemoryCatalog {
	return &InMemoryCatalog{
		messages: make(map[string]map[string]*Message),
	}
}

// AddMessage adds a message to the catalog.
func (c *InMemoryCatalog) AddMessage(locale, key string, msg *Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.messages[locale] == nil {
		c.messages[locale] = make(map[string]*Message)
	}
	c.messages[locale][key] = msg
}

// AddSimpleMessage adds a simple string message to the catalog.
func (c *InMemoryCatalog) AddSimpleMessage(locale, key, text string) {
	c.AddMessage(locale, key, &Message{ID: key, Other: text})
}

// AddPluralMessage adds a plural message to the catalog.
func (c *InMemoryCatalog) AddPluralMessage(locale, key, one, other string) {
	c.AddMessage(locale, key, &Message{ID: key, One: one, Other: other})
}

// Lookup finds a message by locale and key.
func (c *InMemoryCatalog) Lookup(locale, key string) (*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	localeMessages, ok := c.messages[locale]
	if !ok {
		return nil, ErrLocaleNotFound
	}

	msg, ok := localeMessages[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return msg, nil
}

// All returns all messages for a locale.
func (c *InMemoryCatalog) All(locale string) (map[string]*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	localeMessages, ok := c.messages[locale]
	if !ok {
		return nil, ErrLocaleNotFound
	}

	result := make(map[string]*Message, len(localeMessages))
	for k, v := range localeMessages {
		result[k] = v
	}

	return result, nil
}

// Locales returns all available locales.
func (c *InMemoryCatalog) Locales() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	locales := make([]string, 0, len(c.messages))
	for locale := range c.messages {
		locales = append(locales, locale)
	}
	return locales
}

// Reload does nothing for in-memory catalog.
func (c *InMemoryCatalog) Reload() error {
	return nil
}
