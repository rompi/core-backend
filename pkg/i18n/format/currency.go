package format

import (
	"strings"
)

// CurrencyInfo holds information about a currency.
type CurrencyInfo struct {
	Code           string
	Symbol         string
	Name           string
	DecimalDigits  int
	SymbolPosition string // "before" or "after"
}

// currencies contains information about common currencies.
var currencies = map[string]CurrencyInfo{
	"USD": {Code: "USD", Symbol: "$", Name: "US Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"EUR": {Code: "EUR", Symbol: "€", Name: "Euro", DecimalDigits: 2, SymbolPosition: "before"},
	"GBP": {Code: "GBP", Symbol: "£", Name: "British Pound", DecimalDigits: 2, SymbolPosition: "before"},
	"JPY": {Code: "JPY", Symbol: "¥", Name: "Japanese Yen", DecimalDigits: 0, SymbolPosition: "before"},
	"CNY": {Code: "CNY", Symbol: "¥", Name: "Chinese Yuan", DecimalDigits: 2, SymbolPosition: "before"},
	"KRW": {Code: "KRW", Symbol: "₩", Name: "South Korean Won", DecimalDigits: 0, SymbolPosition: "before"},
	"INR": {Code: "INR", Symbol: "₹", Name: "Indian Rupee", DecimalDigits: 2, SymbolPosition: "before"},
	"RUB": {Code: "RUB", Symbol: "₽", Name: "Russian Ruble", DecimalDigits: 2, SymbolPosition: "after"},
	"BRL": {Code: "BRL", Symbol: "R$", Name: "Brazilian Real", DecimalDigits: 2, SymbolPosition: "before"},
	"CAD": {Code: "CAD", Symbol: "CA$", Name: "Canadian Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"AUD": {Code: "AUD", Symbol: "A$", Name: "Australian Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"CHF": {Code: "CHF", Symbol: "CHF", Name: "Swiss Franc", DecimalDigits: 2, SymbolPosition: "before"},
	"HKD": {Code: "HKD", Symbol: "HK$", Name: "Hong Kong Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"SGD": {Code: "SGD", Symbol: "S$", Name: "Singapore Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"SEK": {Code: "SEK", Symbol: "kr", Name: "Swedish Krona", DecimalDigits: 2, SymbolPosition: "after"},
	"NOK": {Code: "NOK", Symbol: "kr", Name: "Norwegian Krone", DecimalDigits: 2, SymbolPosition: "after"},
	"DKK": {Code: "DKK", Symbol: "kr", Name: "Danish Krone", DecimalDigits: 2, SymbolPosition: "after"},
	"PLN": {Code: "PLN", Symbol: "zł", Name: "Polish Zloty", DecimalDigits: 2, SymbolPosition: "after"},
	"MXN": {Code: "MXN", Symbol: "MX$", Name: "Mexican Peso", DecimalDigits: 2, SymbolPosition: "before"},
	"NZD": {Code: "NZD", Symbol: "NZ$", Name: "New Zealand Dollar", DecimalDigits: 2, SymbolPosition: "before"},
	"THB": {Code: "THB", Symbol: "฿", Name: "Thai Baht", DecimalDigits: 2, SymbolPosition: "before"},
	"IDR": {Code: "IDR", Symbol: "Rp", Name: "Indonesian Rupiah", DecimalDigits: 0, SymbolPosition: "before"},
	"TRY": {Code: "TRY", Symbol: "₺", Name: "Turkish Lira", DecimalDigits: 2, SymbolPosition: "before"},
	"SAR": {Code: "SAR", Symbol: "﷼", Name: "Saudi Riyal", DecimalDigits: 2, SymbolPosition: "before"},
	"AED": {Code: "AED", Symbol: "د.إ", Name: "UAE Dirham", DecimalDigits: 2, SymbolPosition: "before"},
	"ZAR": {Code: "ZAR", Symbol: "R", Name: "South African Rand", DecimalDigits: 2, SymbolPosition: "before"},
	"PHP": {Code: "PHP", Symbol: "₱", Name: "Philippine Peso", DecimalDigits: 2, SymbolPosition: "before"},
	"VND": {Code: "VND", Symbol: "₫", Name: "Vietnamese Dong", DecimalDigits: 0, SymbolPosition: "after"},
	"MYR": {Code: "MYR", Symbol: "RM", Name: "Malaysian Ringgit", DecimalDigits: 2, SymbolPosition: "before"},
	"TWD": {Code: "TWD", Symbol: "NT$", Name: "Taiwan Dollar", DecimalDigits: 0, SymbolPosition: "before"},
	"CZK": {Code: "CZK", Symbol: "Kč", Name: "Czech Koruna", DecimalDigits: 2, SymbolPosition: "after"},
	"ILS": {Code: "ILS", Symbol: "₪", Name: "Israeli Shekel", DecimalDigits: 2, SymbolPosition: "before"},
	"CLP": {Code: "CLP", Symbol: "CLP$", Name: "Chilean Peso", DecimalDigits: 0, SymbolPosition: "before"},
	"ARS": {Code: "ARS", Symbol: "AR$", Name: "Argentine Peso", DecimalDigits: 2, SymbolPosition: "before"},
	"COP": {Code: "COP", Symbol: "CO$", Name: "Colombian Peso", DecimalDigits: 0, SymbolPosition: "before"},
	"PEN": {Code: "PEN", Symbol: "S/", Name: "Peruvian Sol", DecimalDigits: 2, SymbolPosition: "before"},
	"EGP": {Code: "EGP", Symbol: "E£", Name: "Egyptian Pound", DecimalDigits: 2, SymbolPosition: "before"},
	"NGN": {Code: "NGN", Symbol: "₦", Name: "Nigerian Naira", DecimalDigits: 2, SymbolPosition: "before"},
	"KES": {Code: "KES", Symbol: "KSh", Name: "Kenyan Shilling", DecimalDigits: 2, SymbolPosition: "before"},
	"UAH": {Code: "UAH", Symbol: "₴", Name: "Ukrainian Hryvnia", DecimalDigits: 2, SymbolPosition: "after"},
	"BGN": {Code: "BGN", Symbol: "лв", Name: "Bulgarian Lev", DecimalDigits: 2, SymbolPosition: "after"},
	"RON": {Code: "RON", Symbol: "lei", Name: "Romanian Leu", DecimalDigits: 2, SymbolPosition: "after"},
	"HRK": {Code: "HRK", Symbol: "kn", Name: "Croatian Kuna", DecimalDigits: 2, SymbolPosition: "after"},
	"HUF": {Code: "HUF", Symbol: "Ft", Name: "Hungarian Forint", DecimalDigits: 0, SymbolPosition: "after"},
}

// localeCurrencyFormats defines currency formatting rules per locale.
type localeCurrencyFormat struct {
	SymbolPosition string // "before", "after", "before_space", "after_space"
	SymbolOverride map[string]string
}

var localeCurrencyFormats = map[string]localeCurrencyFormat{
	"en":    {SymbolPosition: "before"},
	"en-US": {SymbolPosition: "before"},
	"en-GB": {SymbolPosition: "before"},
	"de":    {SymbolPosition: "after_space"},
	"de-DE": {SymbolPosition: "after_space"},
	"fr":    {SymbolPosition: "after_space"},
	"fr-FR": {SymbolPosition: "after_space"},
	"es":    {SymbolPosition: "after_space"},
	"es-ES": {SymbolPosition: "after_space"},
	"it":    {SymbolPosition: "after_space"},
	"it-IT": {SymbolPosition: "after_space"},
	"pt":    {SymbolPosition: "before_space"},
	"pt-BR": {SymbolPosition: "before_space"},
	"nl":    {SymbolPosition: "before_space"},
	"ru":    {SymbolPosition: "after_space"},
	"ja":    {SymbolPosition: "before"},
	"ja-JP": {SymbolPosition: "before"},
	"zh":    {SymbolPosition: "before"},
	"zh-CN": {SymbolPosition: "before"},
	"ko":    {SymbolPosition: "before"},
	"ko-KR": {SymbolPosition: "before"},
	"ar":    {SymbolPosition: "after_space"},
	"he":    {SymbolPosition: "after_space"},
	"hi":    {SymbolPosition: "before"},
	"th":    {SymbolPosition: "before"},
	"vi":    {SymbolPosition: "after_space"},
	"id":    {SymbolPosition: "before"},
	"pl":    {SymbolPosition: "after_space"},
	"tr":    {SymbolPosition: "before"},
	"sv":    {SymbolPosition: "after_space"},
	"da":    {SymbolPosition: "after_space"},
	"no":    {SymbolPosition: "after_space"},
	"fi":    {SymbolPosition: "after_space"},
	"cs":    {SymbolPosition: "after_space"},
	"uk":    {SymbolPosition: "after_space"},
	"ro":    {SymbolPosition: "after_space"},
	"hu":    {SymbolPosition: "after_space"},
	"el":    {SymbolPosition: "after_space"},
}

// GetCurrencyInfo returns information about a currency.
func GetCurrencyInfo(code string) CurrencyInfo {
	code = strings.ToUpper(code)
	if info, ok := currencies[code]; ok {
		return info
	}
	// Return a default for unknown currencies
	return CurrencyInfo{
		Code:           code,
		Symbol:         code,
		Name:           code,
		DecimalDigits:  2,
		SymbolPosition: "before",
	}
}

// FormatCurrency formats a currency amount according to locale conventions.
func FormatCurrency(locale string, amount float64, currency string, cfg FormatConfig) string {
	currencyInfo := GetCurrencyInfo(currency)

	// Override decimal places from currency info
	currencyCfg := cfg
	currencyCfg.MinDecimals = currencyInfo.DecimalDigits
	currencyCfg.MaxDecimals = currencyInfo.DecimalDigits

	// Format the number
	formatted := FormatNumber(locale, amount, currencyCfg)

	// Get the currency display string
	var currencyStr string
	switch cfg.CurrencyDisplay {
	case CurrencyCode:
		currencyStr = currencyInfo.Code
	case CurrencyName:
		currencyStr = currencyInfo.Name
	default: // CurrencySymbol
		currencyStr = currencyInfo.Symbol
	}

	// Get locale-specific formatting
	format := getLocaleCurrencyFormat(locale)

	// Combine currency symbol and amount
	return combineCurrencyAndAmount(formatted, currencyStr, format.SymbolPosition)
}

// getLocaleCurrencyFormat returns the currency format for a locale.
func getLocaleCurrencyFormat(locale string) localeCurrencyFormat {
	// Try exact match
	if fmt, ok := localeCurrencyFormats[locale]; ok {
		return fmt
	}

	// Try language only
	if idx := strings.Index(locale, "-"); idx != -1 {
		lang := locale[:idx]
		if fmt, ok := localeCurrencyFormats[lang]; ok {
			return fmt
		}
	}

	// Default to English
	return localeCurrencyFormats["en"]
}

// combineCurrencyAndAmount combines the currency symbol with the formatted amount.
func combineCurrencyAndAmount(amount, symbol, position string) string {
	switch position {
	case "after":
		return amount + symbol
	case "after_space":
		return amount + " " + symbol
	case "before_space":
		return symbol + " " + amount
	default: // "before"
		return symbol + amount
	}
}
