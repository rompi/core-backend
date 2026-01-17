// Package i18n provides internationalization (i18n) and localization (l10n)
// functionality for building multi-language applications.
package i18n

import "errors"

var (
	// ErrLocaleNotFound is returned when the requested locale is not available.
	ErrLocaleNotFound = errors.New("i18n: locale not found")

	// ErrKeyNotFound is returned when a translation key does not exist.
	ErrKeyNotFound = errors.New("i18n: key not found")

	// ErrInvalidFormat is returned when a translation format is invalid.
	ErrInvalidFormat = errors.New("i18n: invalid format")

	// ErrCatalogLoad is returned when the catalog fails to load.
	ErrCatalogLoad = errors.New("i18n: catalog load failed")

	// ErrInvalidLocale is returned when a locale string is malformed.
	ErrInvalidLocale = errors.New("i18n: invalid locale")

	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("i18n: invalid configuration")

	// ErrTemplateExecution is returned when template execution fails.
	ErrTemplateExecution = errors.New("i18n: template execution failed")
)
