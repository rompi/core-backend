package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// JSONCatalog implements Catalog using JSON files.
type JSONCatalog struct {
	path     string
	messages map[string]map[string]*Message // locale -> key -> message
	locales  []string
	mu       sync.RWMutex
}

// NewJSONCatalog creates a new JSON-based catalog from the given path.
// The path should point to a directory containing JSON files named by locale (e.g., en.json, es.json).
func NewJSONCatalog(path string) (*JSONCatalog, error) {
	c := &JSONCatalog{
		path:     path,
		messages: make(map[string]map[string]*Message),
	}

	if err := c.load(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFailed, err)
	}

	return c, nil
}

// Lookup finds a message by locale and key.
func (c *JSONCatalog) Lookup(locale, key string) (*Message, error) {
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
func (c *JSONCatalog) All(locale string) (map[string]*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	localeMessages, ok := c.messages[locale]
	if !ok {
		return nil, ErrLocaleNotFound
	}

	// Return a copy to prevent external modification
	result := make(map[string]*Message, len(localeMessages))
	for k, v := range localeMessages {
		result[k] = v
	}

	return result, nil
}

// Locales returns all available locales.
func (c *JSONCatalog) Locales() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.locales))
	copy(result, c.locales)
	return result
}

// Reload reloads the catalog from disk.
func (c *JSONCatalog) Reload() error {
	return c.load()
}

// load loads all JSON files from the configured path.
func (c *JSONCatalog) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if path exists
	info, err := os.Stat(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, start with empty catalog
			c.messages = make(map[string]map[string]*Message)
			c.locales = nil
			return nil
		}
		return fmt.Errorf("failed to stat path %s: %w", c.path, err)
	}

	newMessages := make(map[string]map[string]*Message)
	var newLocales []string

	if info.IsDir() {
		// Load all JSON files in directory
		entries, err := os.ReadDir(c.path)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", c.path, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			locale := strings.TrimSuffix(entry.Name(), ".json")
			filePath := filepath.Join(c.path, entry.Name())

			messages, err := c.loadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", filePath, err)
			}

			newMessages[locale] = messages
			newLocales = append(newLocales, locale)
		}
	} else {
		// Single file - extract locale from filename
		locale := strings.TrimSuffix(filepath.Base(c.path), ".json")
		messages, err := c.loadFile(c.path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", c.path, err)
		}

		newMessages[locale] = messages
		newLocales = append(newLocales, locale)
	}

	c.messages = newMessages
	c.locales = newLocales

	return nil
}

// loadFile loads a single JSON file and returns the messages.
func (c *JSONCatalog) loadFile(path string) (map[string]*Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse as generic JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	messages := make(map[string]*Message)
	c.parseMessages("", raw, messages)

	return messages, nil
}

// parseMessages recursively parses JSON data into messages.
func (c *JSONCatalog) parseMessages(prefix string, data map[string]interface{}, messages map[string]*Message) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			// Simple string message
			messages[fullKey] = &Message{
				ID:    fullKey,
				Other: v,
			}

		case map[string]interface{}:
			// Could be plural forms or nested keys
			if c.isPluralObject(v) {
				msg := c.parsePluralMessage(fullKey, v)
				messages[fullKey] = msg
			} else {
				// Nested keys
				c.parseMessages(fullKey, v, messages)
			}
		}
	}
}

// isPluralObject checks if the map contains plural form keys.
func (c *JSONCatalog) isPluralObject(m map[string]interface{}) bool {
	pluralKeys := []string{"one", "other", "zero", "two", "few", "many"}
	for _, pk := range pluralKeys {
		if _, ok := m[pk]; ok {
			return true
		}
	}
	return false
}

// parsePluralMessage creates a Message from a plural object.
func (c *JSONCatalog) parsePluralMessage(key string, m map[string]interface{}) *Message {
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
