package i18n

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"
)

// contextKey is used for context values.
type contextKey string

const (
	localeContextKey contextKey = "i18n_locale"
)

// DateStyle defines the style for date formatting.
type DateStyle int

const (
	// DateStyleShort formats as 1/15/24.
	DateStyleShort DateStyle = iota
	// DateStyleMedium formats as Jan 15, 2024.
	DateStyleMedium
	// DateStyleLong formats as January 15, 2024.
	DateStyleLong
	// DateStyleFull formats as Monday, January 15, 2024.
	DateStyleFull
)

// TimeStyle defines the style for time formatting.
type TimeStyle int

const (
	// TimeStyleShort formats as 3:04 PM.
	TimeStyleShort TimeStyle = iota
	// TimeStyleMedium formats as 3:04:05 PM.
	TimeStyleMedium
	// TimeStyleLong formats as 3:04:05 PM EST.
	TimeStyleLong
)

// ListStyle defines the style for list formatting.
type ListStyle int

const (
	// ListStyleAnd formats as "a, b, and c".
	ListStyleAnd ListStyle = iota
	// ListStyleOr formats as "a, b, or c".
	ListStyleOr
	// ListStyleNarrow formats as "a, b, c".
	ListStyleNarrow
)

// I18n provides internationalization functionality.
type I18n interface {
	// T translates a message key with positional arguments.
	T(ctx context.Context, key string, args ...interface{}) string

	// Tn translates a message key with pluralization.
	Tn(ctx context.Context, key string, count int, args ...interface{}) string

	// Tf translates a message key with named arguments.
	Tf(ctx context.Context, key string, args map[string]interface{}) string

	// L returns a locale-specific localizer.
	L(locale string) Localizer

	// Locale returns the locale from context.
	Locale(ctx context.Context) string

	// WithLocale returns a context with the specified locale.
	WithLocale(ctx context.Context, locale string) context.Context

	// Locales returns all available locales.
	Locales() []string

	// Reload reloads translations from the catalog.
	Reload() error
}

// Localizer provides locale-specific operations.
type Localizer interface {
	// T translates a message key with positional arguments.
	T(key string, args ...interface{}) string

	// Tn translates a message key with pluralization.
	Tn(key string, count int, args ...interface{}) string

	// Tf translates a message key with named arguments.
	Tf(key string, args map[string]interface{}) string

	// FormatNumber formats a number according to locale conventions.
	FormatNumber(n float64, opts ...FormatOption) string

	// FormatCurrency formats a currency amount according to locale conventions.
	FormatCurrency(amount float64, currency string, opts ...FormatOption) string

	// FormatDate formats a date according to locale conventions.
	FormatDate(t time.Time, style DateStyle) string

	// FormatTime formats a time according to locale conventions.
	FormatTime(t time.Time, style TimeStyle) string

	// FormatDateTime formats a date and time according to locale conventions.
	FormatDateTime(t time.Time, dateStyle DateStyle, timeStyle TimeStyle) string

	// FormatRelativeTime formats a time as relative to now (e.g., "2 hours ago").
	FormatRelativeTime(t time.Time) string

	// FormatList formats a list of items according to locale conventions.
	FormatList(items []string, style ListStyle) string

	// FormatPercent formats a number as a percentage.
	FormatPercent(n float64, opts ...FormatOption) string

	// Locale returns the locale identifier.
	Locale() string

	// Language returns the language code.
	Language() string

	// Region returns the region code.
	Region() string

	// Direction returns the text direction (LTR or RTL).
	Direction() Direction
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

// i18nImpl is the default implementation of I18n.
type i18nImpl struct {
	config         *Config
	catalog        Catalog
	logger         Logger
	missingHandler MissingHandler
	localeMatcher  *LocaleMatcher
	localizers     map[string]*localizerImpl
	localizersMu   sync.RWMutex
}

// New creates a new I18n instance with the given configuration and options.
func New(cfg Config, opts ...Option) (I18n, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	impl := &i18nImpl{
		config:     &cfg,
		logger:     NewNoopLogger(),
		localizers: make(map[string]*localizerImpl),
	}

	// Apply options
	for _, opt := range opts {
		opt(impl)
	}

	// Initialize catalog if not provided via options
	if impl.catalog == nil {
		cat, err := createCatalog(cfg)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCatalogLoad, err)
		}
		impl.catalog = cat
	}

	// Initialize locale matcher
	impl.localeMatcher = NewLocaleMatcher(impl.catalog.Locales())

	impl.logger.Info("i18n initialized",
		"default_locale", cfg.DefaultLocale,
		"fallback_locale", cfg.FallbackLocale,
		"available_locales", impl.catalog.Locales(),
	)

	return impl, nil
}

// createCatalog creates a catalog based on configuration.
func createCatalog(cfg Config) (Catalog, error) {
	switch cfg.CatalogType {
	case CatalogTypeJSON:
		return NewJSONCatalog(cfg.Path)
	case CatalogTypeYAML:
		return NewYAMLCatalog(cfg.Path)
	default:
		return NewJSONCatalog(cfg.Path)
	}
}

// T translates a message key with positional arguments.
func (i *i18nImpl) T(ctx context.Context, key string, args ...interface{}) string {
	locale := i.resolveLocale(ctx)
	return i.getLocalizer(locale).T(key, args...)
}

// Tn translates a message key with pluralization.
func (i *i18nImpl) Tn(ctx context.Context, key string, count int, args ...interface{}) string {
	locale := i.resolveLocale(ctx)
	return i.getLocalizer(locale).Tn(key, count, args...)
}

// Tf translates a message key with named arguments.
func (i *i18nImpl) Tf(ctx context.Context, key string, args map[string]interface{}) string {
	locale := i.resolveLocale(ctx)
	return i.getLocalizer(locale).Tf(key, args)
}

// L returns a locale-specific localizer.
func (i *i18nImpl) L(locale string) Localizer {
	return i.getLocalizer(locale)
}

// Locale returns the locale from context.
func (i *i18nImpl) Locale(ctx context.Context) string {
	return i.resolveLocale(ctx)
}

// WithLocale returns a context with the specified locale.
func (i *i18nImpl) WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey, locale)
}

// Locales returns all available locales.
func (i *i18nImpl) Locales() []string {
	return i.catalog.Locales()
}

// Reload reloads translations from the catalog.
func (i *i18nImpl) Reload() error {
	if err := i.catalog.Reload(); err != nil {
		return err
	}

	// Clear cached localizers
	i.localizersMu.Lock()
	i.localizers = make(map[string]*localizerImpl)
	i.localizersMu.Unlock()

	// Update locale matcher
	i.localeMatcher = NewLocaleMatcher(i.catalog.Locales())

	i.logger.Info("i18n reloaded", "available_locales", i.catalog.Locales())
	return nil
}

// resolveLocale resolves the locale from context, with fallbacks.
func (i *i18nImpl) resolveLocale(ctx context.Context) string {
	// Try to get locale from context
	if v := ctx.Value(localeContextKey); v != nil {
		if locale, ok := v.(string); ok && locale != "" {
			// Try to match the requested locale
			if matched := i.localeMatcher.Match(locale); matched != "" {
				return matched
			}
		}
	}

	// Fall back to default locale
	return i.config.DefaultLocale
}

// getLocalizer returns a localizer for the given locale, creating one if necessary.
func (i *i18nImpl) getLocalizer(locale string) *localizerImpl {
	i.localizersMu.RLock()
	if l, ok := i.localizers[locale]; ok {
		i.localizersMu.RUnlock()
		return l
	}
	i.localizersMu.RUnlock()

	// Create new localizer
	i.localizersMu.Lock()
	defer i.localizersMu.Unlock()

	// Double-check after acquiring write lock
	if l, ok := i.localizers[locale]; ok {
		return l
	}

	parsed, _ := ParseLocale(locale)
	if parsed == nil {
		parsed = &Locale{Language: locale}
	}

	l := &localizerImpl{
		i18n:       i,
		locale:     locale,
		parsedLoc:  parsed,
		pluralRule: GetPluralRule(parsed.Language),
	}

	i.localizers[locale] = l
	return l
}

// localizerImpl is the default implementation of Localizer.
type localizerImpl struct {
	i18n       *i18nImpl
	locale     string
	parsedLoc  *Locale
	pluralRule PluralRule
}

// T translates a message key with positional arguments.
func (l *localizerImpl) T(key string, args ...interface{}) string {
	msg, err := l.lookupMessage(key)
	if err != nil {
		return l.handleMissing(key)
	}

	text := msg.SimpleMessage()
	return l.interpolate(text, args)
}

// Tn translates a message key with pluralization.
func (l *localizerImpl) Tn(key string, count int, args ...interface{}) string {
	msg, err := l.lookupMessage(key)
	if err != nil {
		return l.handleMissing(key)
	}

	category := l.pluralRule(count)
	text := msg.GetForm(category)

	// Prepend count to args for interpolation
	allArgs := append([]interface{}{count}, args...)
	return l.interpolate(text, allArgs)
}

// Tf translates a message key with named arguments.
func (l *localizerImpl) Tf(key string, args map[string]interface{}) string {
	msg, err := l.lookupMessage(key)
	if err != nil {
		return l.handleMissing(key)
	}

	text := msg.SimpleMessage()
	return l.interpolateNamed(text, args)
}

// lookupMessage looks up a message, trying locale chain.
func (l *localizerImpl) lookupMessage(key string) (*Message, error) {
	// Try current locale
	msg, err := l.i18n.catalog.Lookup(l.locale, key)
	if err == nil && msg != nil {
		return msg, nil
	}

	// Try language only (strip region)
	if l.parsedLoc != nil && l.parsedLoc.Region != "" {
		msg, err = l.i18n.catalog.Lookup(l.parsedLoc.Language, key)
		if err == nil && msg != nil {
			return msg, nil
		}
	}

	// Try fallback locale
	if l.locale != l.i18n.config.FallbackLocale {
		msg, err = l.i18n.catalog.Lookup(l.i18n.config.FallbackLocale, key)
		if err == nil && msg != nil {
			return msg, nil
		}
	}

	return nil, ErrKeyNotFound
}

// handleMissing handles a missing translation.
func (l *localizerImpl) handleMissing(key string) string {
	if l.i18n.config.LogMissing {
		l.i18n.logger.Warn("missing translation", "locale", l.locale, "key", key)
	}

	if l.i18n.missingHandler != nil {
		l.i18n.missingHandler(l.locale, key)
	}

	switch l.i18n.config.MissingKeyBehavior {
	case MissingKeyReturnEmpty:
		return ""
	case MissingKeyReturnError:
		return fmt.Sprintf("[MISSING: %s]", key)
	default:
		return key
	}
}

// interpolate interpolates positional arguments into the text.
func (l *localizerImpl) interpolate(text string, args []interface{}) string {
	if len(args) == 0 {
		return text
	}

	// If template contains Go template syntax, use template engine
	if strings.Contains(text, "{{") {
		data := make(map[string]interface{})
		if len(args) > 0 {
			data["Count"] = args[0]
		}
		for i, arg := range args {
			data[fmt.Sprintf("Arg%d", i)] = arg
		}
		return l.interpolateNamed(text, data)
	}

	// Simple printf-style interpolation
	return fmt.Sprintf(text, args...)
}

// interpolateNamed interpolates named arguments into the text using Go templates.
func (l *localizerImpl) interpolateNamed(text string, args map[string]interface{}) string {
	if args == nil || !strings.Contains(text, "{{") {
		return text
	}

	tmpl, err := template.New("msg").Parse(text)
	if err != nil {
		l.i18n.logger.Error("template parse error", "text", text, "error", err)
		return text
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		l.i18n.logger.Error("template execute error", "text", text, "error", err)
		return text
	}

	return buf.String()
}

// Locale returns the locale identifier.
func (l *localizerImpl) Locale() string {
	return l.locale
}

// Language returns the language code.
func (l *localizerImpl) Language() string {
	if l.parsedLoc != nil {
		return l.parsedLoc.Language
	}
	return l.locale
}

// Region returns the region code.
func (l *localizerImpl) Region() string {
	if l.parsedLoc != nil {
		return l.parsedLoc.Region
	}
	return ""
}

// Direction returns the text direction.
func (l *localizerImpl) Direction() Direction {
	if l.parsedLoc != nil {
		return l.parsedLoc.Direction()
	}
	return LTR
}

// Formatting methods are implemented in format/*.go files
// The following are placeholder implementations that call into the format package.

// FormatNumber formats a number according to locale conventions.
func (l *localizerImpl) FormatNumber(n float64, opts ...FormatOption) string {
	return formatNumber(l.locale, n, opts...)
}

// FormatCurrency formats a currency amount according to locale conventions.
func (l *localizerImpl) FormatCurrency(amount float64, currency string, opts ...FormatOption) string {
	return formatCurrency(l.locale, amount, currency, opts...)
}

// FormatDate formats a date according to locale conventions.
func (l *localizerImpl) FormatDate(t time.Time, style DateStyle) string {
	return formatDate(l.locale, t, style)
}

// FormatTime formats a time according to locale conventions.
func (l *localizerImpl) FormatTime(t time.Time, style TimeStyle) string {
	return formatTime(l.locale, t, style)
}

// FormatDateTime formats a date and time according to locale conventions.
func (l *localizerImpl) FormatDateTime(t time.Time, dateStyle DateStyle, timeStyle TimeStyle) string {
	return formatDateTime(l.locale, t, dateStyle, timeStyle)
}

// FormatRelativeTime formats a time as relative to now.
func (l *localizerImpl) FormatRelativeTime(t time.Time) string {
	return formatRelativeTime(l.locale, t)
}

// FormatList formats a list of items according to locale conventions.
func (l *localizerImpl) FormatList(items []string, style ListStyle) string {
	return formatList(l.locale, items, style)
}

// FormatPercent formats a number as a percentage.
func (l *localizerImpl) FormatPercent(n float64, opts ...FormatOption) string {
	return formatPercent(l.locale, n, opts...)
}
