package i18n

// Option is a functional option for configuring an I18n instance.
type Option func(*i18nImpl)

// WithLogger sets the logger for the i18n instance.
func WithLogger(logger Logger) Option {
	return func(i *i18nImpl) {
		if logger != nil {
			i.logger = logger
		}
	}
}

// WithCatalog sets a custom catalog for the i18n instance.
func WithCatalog(catalog Catalog) Option {
	return func(i *i18nImpl) {
		if catalog != nil {
			i.catalog = catalog
		}
	}
}

// WithMissingHandler sets a custom handler for missing translations.
func WithMissingHandler(handler MissingHandler) Option {
	return func(i *i18nImpl) {
		if handler != nil {
			i.missingHandler = handler
		}
	}
}

// MissingHandler is a function called when a translation is missing.
type MissingHandler func(locale, key string)

// FormatOption is a functional option for configuring formatting.
type FormatOption func(*formatConfig)

// formatConfig holds formatting configuration.
type formatConfig struct {
	minDecimals     int
	maxDecimals     int
	useGrouping     bool
	currencyDisplay CurrencyDisplay
}

// defaultFormatConfig returns the default formatting configuration.
func defaultFormatConfig() *formatConfig {
	return &formatConfig{
		minDecimals:     0,
		maxDecimals:     3,
		useGrouping:     true,
		currencyDisplay: CurrencySymbol,
	}
}

// WithMinDecimals sets the minimum number of decimal places.
func WithMinDecimals(n int) FormatOption {
	return func(c *formatConfig) {
		if n >= 0 {
			c.minDecimals = n
		}
	}
}

// WithMaxDecimals sets the maximum number of decimal places.
func WithMaxDecimals(n int) FormatOption {
	return func(c *formatConfig) {
		if n >= 0 {
			c.maxDecimals = n
		}
	}
}

// WithGrouping enables or disables thousand separators.
func WithGrouping(enabled bool) FormatOption {
	return func(c *formatConfig) {
		c.useGrouping = enabled
	}
}

// WithCurrencyDisplay sets the currency display mode.
func WithCurrencyDisplay(display CurrencyDisplay) FormatOption {
	return func(c *formatConfig) {
		c.currencyDisplay = display
	}
}

// CurrencyDisplay defines how currency is displayed.
type CurrencyDisplay int

const (
	// CurrencySymbol displays the currency symbol (e.g., $).
	CurrencySymbol CurrencyDisplay = iota

	// CurrencyCode displays the currency code (e.g., USD).
	CurrencyCode

	// CurrencyName displays the currency name (e.g., US Dollar).
	CurrencyName
)
