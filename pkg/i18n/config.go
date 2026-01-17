package i18n

import (
	"fmt"
	"os"
	"strings"
)

// MissingKeyBehavior defines how missing translation keys are handled.
type MissingKeyBehavior string

const (
	// MissingKeyReturnKey returns the key itself when translation is missing.
	MissingKeyReturnKey MissingKeyBehavior = "key"

	// MissingKeyReturnEmpty returns an empty string when translation is missing.
	MissingKeyReturnEmpty MissingKeyBehavior = "empty"

	// MissingKeyReturnError returns an error marker when translation is missing.
	MissingKeyReturnError MissingKeyBehavior = "error"
)

// CatalogType defines the type of translation catalog.
type CatalogType string

const (
	// CatalogTypeJSON uses JSON files for translations.
	CatalogTypeJSON CatalogType = "json"

	// CatalogTypeYAML uses YAML files for translations.
	CatalogTypeYAML CatalogType = "yaml"

	// CatalogTypePO uses Gettext PO/MO files for translations.
	CatalogTypePO CatalogType = "po"

	// CatalogTypeDatabase uses a database for translations.
	CatalogTypeDatabase CatalogType = "database"

	// CatalogTypeEmbed uses embedded filesystem for translations.
	CatalogTypeEmbed CatalogType = "embed"

	// CatalogTypeCustom indicates a custom catalog will be provided via WithCatalog option.
	CatalogTypeCustom CatalogType = "custom"
)

// Config holds the configuration for the i18n instance.
type Config struct {
	// DefaultLocale is the default locale used when no locale is specified.
	// Environment variable: I18N_DEFAULT_LOCALE
	// Default: "en"
	DefaultLocale string `json:"default_locale"`

	// FallbackLocale is the locale used when a translation is missing in the requested locale.
	// Environment variable: I18N_FALLBACK_LOCALE
	// Default: "en"
	FallbackLocale string `json:"fallback_locale"`

	// CatalogType specifies the type of translation catalog to use.
	// Environment variable: I18N_CATALOG_TYPE
	// Default: "json"
	CatalogType CatalogType `json:"catalog_type"`

	// Path is the path to translation files.
	// Environment variable: I18N_PATH
	// Default: "./locales"
	Path string `json:"path"`

	// HotReload enables automatic reloading of translations when files change.
	// Environment variable: I18N_HOT_RELOAD
	// Default: false
	HotReload bool `json:"hot_reload"`

	// MissingKeyBehavior defines how missing translation keys are handled.
	// Environment variable: I18N_MISSING_KEY
	// Default: "key"
	MissingKeyBehavior MissingKeyBehavior `json:"missing_key_behavior"`

	// LogMissing enables logging of missing translations.
	// Environment variable: I18N_LOG_MISSING
	// Default: true
	LogMissing bool `json:"log_missing"`
}

// LoadConfig loads configuration from environment variables with sensible defaults.
func LoadConfig() Config {
	cfg := defaultConfig()
	cfg.overrideFromEnv()
	return cfg
}

// defaultConfig returns a Config with sensible default values.
func defaultConfig() Config {
	return Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		CatalogType:        CatalogTypeJSON,
		Path:               "./locales",
		HotReload:          false,
		MissingKeyBehavior: MissingKeyReturnKey,
		LogMissing:         true,
	}
}

// overrideFromEnv overrides configuration values from environment variables.
func (c *Config) overrideFromEnv() {
	if v := os.Getenv("I18N_DEFAULT_LOCALE"); v != "" {
		c.DefaultLocale = v
	}

	if v := os.Getenv("I18N_FALLBACK_LOCALE"); v != "" {
		c.FallbackLocale = v
	}

	if v := os.Getenv("I18N_CATALOG_TYPE"); v != "" {
		c.CatalogType = CatalogType(strings.ToLower(v))
	}

	if v := os.Getenv("I18N_PATH"); v != "" {
		c.Path = v
	}

	if v := os.Getenv("I18N_HOT_RELOAD"); v != "" {
		c.HotReload = parseBool(v)
	}

	if v := os.Getenv("I18N_MISSING_KEY"); v != "" {
		c.MissingKeyBehavior = MissingKeyBehavior(strings.ToLower(v))
	}

	if v := os.Getenv("I18N_LOG_MISSING"); v != "" {
		c.LogMissing = parseBool(v)
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.DefaultLocale == "" {
		return fmt.Errorf("%w: default_locale cannot be empty", ErrInvalidConfig)
	}

	if c.FallbackLocale == "" {
		return fmt.Errorf("%w: fallback_locale cannot be empty", ErrInvalidConfig)
	}

	switch c.CatalogType {
	case CatalogTypeJSON, CatalogTypeYAML, CatalogTypePO, CatalogTypeDatabase, CatalogTypeEmbed, CatalogTypeCustom, "":
		// Valid catalog types (empty is allowed when catalog is provided via WithCatalog)
	default:
		return fmt.Errorf("%w: invalid catalog_type: %s", ErrInvalidConfig, c.CatalogType)
	}

	switch c.MissingKeyBehavior {
	case MissingKeyReturnKey, MissingKeyReturnEmpty, MissingKeyReturnError:
		// Valid missing key behaviors
	default:
		return fmt.Errorf("%w: invalid missing_key_behavior: %s", ErrInvalidConfig, c.MissingKeyBehavior)
	}

	return nil
}

// parseBool parses a string as a boolean value.
func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}
