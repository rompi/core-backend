package i18n

import (
	"github.com/rompi/core-backend/pkg/i18n/catalog"
)

// NewJSONCatalog creates a new JSON-based catalog from the given path.
func NewJSONCatalog(path string) (Catalog, error) {
	cat, err := catalog.NewJSONCatalog(path)
	if err != nil {
		return nil, err
	}
	return &catalogAdapter{cat: cat}, nil
}

// NewYAMLCatalog creates a new YAML-based catalog from the given path.
func NewYAMLCatalog(path string) (Catalog, error) {
	cat, err := catalog.NewYAMLCatalog(path)
	if err != nil {
		return nil, err
	}
	return &catalogAdapter{cat: cat}, nil
}

// catalogAdapter wraps a catalog.Catalog to implement i18n.Catalog.
type catalogAdapter struct {
	cat catalog.Catalog
}

// Lookup finds a message by locale and key.
func (a *catalogAdapter) Lookup(locale, key string) (*Message, error) {
	msg, err := a.cat.Lookup(locale, key)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:          msg.ID,
		Description: msg.Description,
		One:         msg.One,
		Other:       msg.Other,
		Zero:        msg.Zero,
		Two:         msg.Two,
		Few:         msg.Few,
		Many:        msg.Many,
	}, nil
}

// All returns all messages for a locale.
func (a *catalogAdapter) All(locale string) (map[string]*Message, error) {
	msgs, err := a.cat.All(locale)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Message, len(msgs))
	for k, msg := range msgs {
		result[k] = &Message{
			ID:          msg.ID,
			Description: msg.Description,
			One:         msg.One,
			Other:       msg.Other,
			Zero:        msg.Zero,
			Two:         msg.Two,
			Few:         msg.Few,
			Many:        msg.Many,
		}
	}
	return result, nil
}

// Locales returns all available locales.
func (a *catalogAdapter) Locales() []string {
	return a.cat.Locales()
}

// Reload reloads the catalog.
func (a *catalogAdapter) Reload() error {
	return a.cat.Reload()
}
