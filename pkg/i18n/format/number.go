// Package format provides locale-aware formatting utilities for numbers,
// currencies, dates, and other values.
package format

import (
	"math"
	"strconv"
	"strings"
)

// NumberFormat holds locale-specific number formatting rules.
type NumberFormat struct {
	DecimalSeparator  string
	GroupingSeparator string
	GroupingSize      int
	MinusSign         string
	PlusSign          string
	PercentSign       string
}

// localeNumberFormats contains number formatting rules for various locales.
var localeNumberFormats = map[string]NumberFormat{
	"en": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"de": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"fr": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"es": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"it": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"pt": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"nl": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"ru": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"ja": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"zh": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"ko": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"ar": {DecimalSeparator: "٫", GroupingSeparator: "٬", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "٪"},
	"he": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"hi": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"th": {DecimalSeparator: ".", GroupingSeparator: ",", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"vi": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"id": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"pl": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"tr": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"sv": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"da": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"no": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"fi": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"cs": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"uk": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"ro": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"hu": {DecimalSeparator: ",", GroupingSeparator: " ", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
	"el": {DecimalSeparator: ",", GroupingSeparator: ".", GroupingSize: 3, MinusSign: "-", PlusSign: "+", PercentSign: "%"},
}

// GetNumberFormat returns the number format for a locale.
func GetNumberFormat(locale string) NumberFormat {
	// Try exact match
	if fmt, ok := localeNumberFormats[locale]; ok {
		return fmt
	}

	// Try language only
	if idx := strings.Index(locale, "-"); idx != -1 {
		lang := locale[:idx]
		if fmt, ok := localeNumberFormats[lang]; ok {
			return fmt
		}
	}

	// Default to English
	return localeNumberFormats["en"]
}

// FormatConfig holds number formatting configuration.
type FormatConfig struct {
	MinDecimals     int
	MaxDecimals     int
	UseGrouping     bool
	CurrencyDisplay CurrencyDisplay
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

// DefaultFormatConfig returns the default format configuration.
func DefaultFormatConfig() FormatConfig {
	return FormatConfig{
		MinDecimals:     0,
		MaxDecimals:     3,
		UseGrouping:     true,
		CurrencyDisplay: CurrencySymbol,
	}
}

// FormatNumber formats a number according to locale conventions.
func FormatNumber(locale string, n float64, cfg FormatConfig) string {
	nf := GetNumberFormat(locale)

	// Handle negative numbers
	negative := n < 0
	if negative {
		n = -n
	}

	// Round to max decimals
	if cfg.MaxDecimals >= 0 {
		multiplier := math.Pow(10, float64(cfg.MaxDecimals))
		n = math.Round(n*multiplier) / multiplier
	}

	// Split into integer and decimal parts
	intPart := int64(n)
	decPart := n - float64(intPart)

	// Format integer part with grouping
	intStr := strconv.FormatInt(intPart, 10)
	if cfg.UseGrouping && nf.GroupingSize > 0 {
		intStr = addGrouping(intStr, nf.GroupingSeparator, nf.GroupingSize)
	}

	// Format decimal part
	var result strings.Builder
	if negative {
		result.WriteString(nf.MinusSign)
	}
	result.WriteString(intStr)

	// Add decimal part if needed
	decStr := formatDecimals(decPart, cfg.MinDecimals, cfg.MaxDecimals)
	if decStr != "" {
		result.WriteString(nf.DecimalSeparator)
		result.WriteString(decStr)
	}

	return result.String()
}

// addGrouping adds thousand separators to an integer string.
func addGrouping(s string, separator string, size int) string {
	if len(s) <= size {
		return s
	}

	var result strings.Builder
	remainder := len(s) % size
	if remainder > 0 {
		result.WriteString(s[:remainder])
		s = s[remainder:]
	}

	for len(s) > 0 {
		if result.Len() > 0 {
			result.WriteString(separator)
		}
		result.WriteString(s[:size])
		s = s[size:]
	}

	return result.String()
}

// formatDecimals formats the decimal part of a number.
func formatDecimals(decPart float64, minDecimals, maxDecimals int) string {
	if maxDecimals <= 0 && minDecimals <= 0 {
		return ""
	}

	// Convert decimal to string
	decStr := strconv.FormatFloat(decPart, 'f', maxDecimals, 64)

	// Remove leading "0."
	if strings.HasPrefix(decStr, "0.") {
		decStr = decStr[2:]
	} else if decStr == "0" {
		decStr = ""
	}

	// Trim trailing zeros up to minDecimals
	for len(decStr) > minDecimals && decStr[len(decStr)-1] == '0' {
		decStr = decStr[:len(decStr)-1]
	}

	// Pad with zeros if needed
	for len(decStr) < minDecimals {
		decStr += "0"
	}

	return decStr
}

// FormatPercent formats a number as a percentage.
func FormatPercent(locale string, n float64, cfg FormatConfig) string {
	nf := GetNumberFormat(locale)
	// Multiply by 100 for percentage
	percentValue := n * 100
	formatted := FormatNumber(locale, percentValue, cfg)
	return formatted + nf.PercentSign
}
