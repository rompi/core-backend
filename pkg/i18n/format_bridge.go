package i18n

import (
	"time"

	"github.com/rompi/core-backend/pkg/i18n/format"
)

// formatNumber formats a number according to locale conventions.
func formatNumber(locale string, n float64, opts ...FormatOption) string {
	cfg := defaultFormatConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	fmtCfg := format.FormatConfig{
		MinDecimals: cfg.minDecimals,
		MaxDecimals: cfg.maxDecimals,
		UseGrouping: cfg.useGrouping,
	}

	return format.FormatNumber(locale, n, fmtCfg)
}

// formatCurrency formats a currency amount according to locale conventions.
func formatCurrency(locale string, amount float64, currency string, opts ...FormatOption) string {
	cfg := defaultFormatConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	fmtCfg := format.FormatConfig{
		MinDecimals:     cfg.minDecimals,
		MaxDecimals:     cfg.maxDecimals,
		UseGrouping:     cfg.useGrouping,
		CurrencyDisplay: format.CurrencyDisplay(cfg.currencyDisplay),
	}

	return format.FormatCurrency(locale, amount, currency, fmtCfg)
}

// formatDate formats a date according to locale conventions.
func formatDate(locale string, t time.Time, style DateStyle) string {
	return format.FormatDate(locale, t, format.DateStyle(style))
}

// formatTime formats a time according to locale conventions.
func formatTime(locale string, t time.Time, style TimeStyle) string {
	return format.FormatTime(locale, t, format.TimeStyle(style))
}

// formatDateTime formats a date and time according to locale conventions.
func formatDateTime(locale string, t time.Time, dateStyle DateStyle, timeStyle TimeStyle) string {
	return format.FormatDateTime(locale, t, format.DateStyle(dateStyle), format.TimeStyle(timeStyle))
}

// formatRelativeTime formats a time as relative to now.
func formatRelativeTime(locale string, t time.Time) string {
	return format.FormatRelativeTime(locale, t)
}

// formatList formats a list of items according to locale conventions.
func formatList(locale string, items []string, style ListStyle) string {
	return format.FormatList(locale, items, format.ListStyle(style))
}

// formatPercent formats a number as a percentage.
func formatPercent(locale string, n float64, opts ...FormatOption) string {
	cfg := defaultFormatConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	fmtCfg := format.FormatConfig{
		MinDecimals: cfg.minDecimals,
		MaxDecimals: cfg.maxDecimals,
		UseGrouping: cfg.useGrouping,
	}

	return format.FormatPercent(locale, n, fmtCfg)
}
