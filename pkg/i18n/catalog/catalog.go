// Package catalog provides message catalog implementations for i18n.
package catalog

import (
	"errors"
)

var (
	// ErrKeyNotFound is returned when a translation key does not exist.
	ErrKeyNotFound = errors.New("catalog: key not found")

	// ErrLocaleNotFound is returned when the requested locale is not available.
	ErrLocaleNotFound = errors.New("catalog: locale not found")

	// ErrLoadFailed is returned when the catalog fails to load.
	ErrLoadFailed = errors.New("catalog: load failed")
)

// Message represents a translatable message with optional plural forms.
type Message struct {
	// ID is the unique identifier for the message.
	ID string `json:"id,omitempty"`

	// Description provides context for translators.
	Description string `json:"description,omitempty"`

	// One is the singular form of the message.
	One string `json:"one,omitempty"`

	// Other is the plural form of the message.
	Other string `json:"other,omitempty"`

	// Zero is the form used when count is zero.
	Zero string `json:"zero,omitempty"`

	// Two is the form used when count is two.
	Two string `json:"two,omitempty"`

	// Few is the form used for small numbers.
	Few string `json:"few,omitempty"`

	// Many is the form used for large numbers.
	Many string `json:"many,omitempty"`
}

// Catalog provides message storage and retrieval.
type Catalog interface {
	// Lookup finds a message by locale and key.
	Lookup(locale, key string) (*Message, error)

	// All returns all messages for a locale.
	All(locale string) (map[string]*Message, error)

	// Locales returns all available locales.
	Locales() []string

	// Reload reloads the catalog from the source.
	Reload() error
}
