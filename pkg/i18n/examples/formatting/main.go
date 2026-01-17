// Package main demonstrates formatting capabilities of the i18n package.
package main

import (
	"fmt"
	"time"

	"github.com/rompi/core-backend/pkg/i18n"
)

func main() {
	// Create i18n instance (no catalog needed for formatting)
	i, err := i18n.New(i18n.Config{
		DefaultLocale: "en",
		Path:          "./locales", // Empty is fine for formatting-only usage
	})
	if err != nil {
		fmt.Println("Note: No locale files found, but formatting still works")
	}

	// Get localizers for different locales
	en := i.L("en-US")
	de := i.L("de-DE")
	fr := i.L("fr-FR")
	ja := i.L("ja-JP")

	// Number formatting
	fmt.Println("=== Number Formatting ===")
	n := 1234567.89
	fmt.Printf("en-US: %s\n", en.FormatNumber(n))
	fmt.Printf("de-DE: %s\n", de.FormatNumber(n))
	fmt.Printf("fr-FR: %s\n", fr.FormatNumber(n))
	fmt.Printf("ja-JP: %s\n", ja.FormatNumber(n))

	// With options
	fmt.Println("\n=== Number Formatting with Options ===")
	fmt.Printf("2 decimals: %s\n", en.FormatNumber(n, i18n.WithMinDecimals(2), i18n.WithMaxDecimals(2)))
	fmt.Printf("No grouping: %s\n", en.FormatNumber(n, i18n.WithGrouping(false)))

	// Currency formatting
	fmt.Println("\n=== Currency Formatting ===")
	amount := 1234.56
	fmt.Printf("USD (en-US): %s\n", en.FormatCurrency(amount, "USD"))
	fmt.Printf("EUR (de-DE): %s\n", de.FormatCurrency(amount, "EUR"))
	fmt.Printf("JPY (ja-JP): %s\n", ja.FormatCurrency(1234, "JPY"))

	// Currency display options
	fmt.Println("\n=== Currency Display Options ===")
	fmt.Printf("Symbol: %s\n", en.FormatCurrency(amount, "USD"))
	fmt.Printf("Code: %s\n", en.FormatCurrency(amount, "USD", i18n.WithCurrencyDisplay(i18n.CurrencyCode)))
	fmt.Printf("Name: %s\n", en.FormatCurrency(amount, "USD", i18n.WithCurrencyDisplay(i18n.CurrencyName)))

	// Date formatting
	fmt.Println("\n=== Date Formatting ===")
	t := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)

	fmt.Println("en-US:")
	fmt.Printf("  Short:  %s\n", en.FormatDate(t, i18n.DateStyleShort))
	fmt.Printf("  Medium: %s\n", en.FormatDate(t, i18n.DateStyleMedium))
	fmt.Printf("  Long:   %s\n", en.FormatDate(t, i18n.DateStyleLong))
	fmt.Printf("  Full:   %s\n", en.FormatDate(t, i18n.DateStyleFull))

	fmt.Println("de-DE:")
	fmt.Printf("  Medium: %s\n", de.FormatDate(t, i18n.DateStyleMedium))

	fmt.Println("ja-JP:")
	fmt.Printf("  Medium: %s\n", ja.FormatDate(t, i18n.DateStyleMedium))

	// Time formatting
	fmt.Println("\n=== Time Formatting ===")
	fmt.Printf("en-US Short:  %s\n", en.FormatTime(t, i18n.TimeStyleShort))
	fmt.Printf("en-US Medium: %s\n", en.FormatTime(t, i18n.TimeStyleMedium))
	fmt.Printf("de-DE Short:  %s\n", de.FormatTime(t, i18n.TimeStyleShort))

	// DateTime formatting
	fmt.Println("\n=== DateTime Formatting ===")
	fmt.Printf("en-US: %s\n", en.FormatDateTime(t, i18n.DateStyleMedium, i18n.TimeStyleShort))
	fmt.Printf("de-DE: %s\n", de.FormatDateTime(t, i18n.DateStyleMedium, i18n.TimeStyleShort))

	// Relative time formatting
	fmt.Println("\n=== Relative Time Formatting ===")
	now := time.Now()
	fmt.Printf("just now:      %s\n", en.FormatRelativeTime(now.Add(-5*time.Second)))
	fmt.Printf("2 minutes ago: %s\n", en.FormatRelativeTime(now.Add(-2*time.Minute)))
	fmt.Printf("3 hours ago:   %s\n", en.FormatRelativeTime(now.Add(-3*time.Hour)))
	fmt.Printf("yesterday:     %s\n", en.FormatRelativeTime(now.Add(-24*time.Hour)))
	fmt.Printf("in 2 hours:    %s\n", en.FormatRelativeTime(now.Add(2*time.Hour)))
	fmt.Printf("tomorrow:      %s\n", en.FormatRelativeTime(now.Add(24*time.Hour)))

	// List formatting
	fmt.Println("\n=== List Formatting ===")
	items := []string{"apples", "oranges", "bananas"}
	fmt.Printf("And (en-US): %s\n", en.FormatList(items, i18n.ListStyleAnd))
	fmt.Printf("Or (en-US):  %s\n", en.FormatList(items, i18n.ListStyleOr))
	fmt.Printf("Narrow:      %s\n", en.FormatList(items, i18n.ListStyleNarrow))

	es := i.L("es")
	fmt.Printf("And (es):    %s\n", es.FormatList(items, i18n.ListStyleAnd))

	// Percent formatting
	fmt.Println("\n=== Percent Formatting ===")
	fmt.Printf("en-US: %s\n", en.FormatPercent(0.1234))
	fmt.Printf("de-DE: %s\n", de.FormatPercent(0.1234))

	// Text direction
	fmt.Println("\n=== Text Direction ===")
	ar := i.L("ar")
	he := i.L("he")
	fmt.Printf("English: %s\n", en.Direction())
	fmt.Printf("Arabic:  %s\n", ar.Direction())
	fmt.Printf("Hebrew:  %s\n", he.Direction())
}
