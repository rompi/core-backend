package format

import (
	"testing"
	"time"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		number float64
		cfg    FormatConfig
		want   string
	}{
		{
			name:   "en-US basic",
			locale: "en-US",
			number: 1234567.89,
			cfg:    DefaultFormatConfig(),
			want:   "1,234,567.89",
		},
		{
			name:   "de-DE basic",
			locale: "de-DE",
			number: 1234567.89,
			cfg:    DefaultFormatConfig(),
			want:   "1.234.567,89",
		},
		{
			name:   "fr-FR basic",
			locale: "fr-FR",
			number: 1234567.89,
			cfg:    DefaultFormatConfig(),
			want:   "1 234 567,89",
		},
		{
			name:   "no grouping",
			locale: "en-US",
			number: 1234567.89,
			cfg:    FormatConfig{UseGrouping: false, MaxDecimals: 2},
			want:   "1234567.89",
		},
		{
			name:   "fixed decimals",
			locale: "en-US",
			number: 1234.5,
			cfg:    FormatConfig{UseGrouping: true, MinDecimals: 2, MaxDecimals: 2},
			want:   "1,234.50",
		},
		{
			name:   "negative number",
			locale: "en-US",
			number: -1234.56,
			cfg:    FormatConfig{UseGrouping: true, MaxDecimals: 2},
			want:   "-1,234.56",
		},
		{
			name:   "zero",
			locale: "en-US",
			number: 0,
			cfg:    DefaultFormatConfig(),
			want:   "0",
		},
		{
			name:   "small decimal",
			locale: "en-US",
			number: 0.123,
			cfg:    FormatConfig{MaxDecimals: 3},
			want:   "0.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatNumber(tt.locale, tt.number, tt.cfg)
			if got != tt.want {
				t.Errorf("FormatNumber() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		amount   float64
		currency string
		cfg      FormatConfig
		want     string
	}{
		{
			name:     "USD en-US",
			locale:   "en-US",
			amount:   1234.56,
			currency: "USD",
			cfg:      DefaultFormatConfig(),
			want:     "$1,234.56",
		},
		{
			name:     "EUR de-DE",
			locale:   "de-DE",
			amount:   1234.56,
			currency: "EUR",
			cfg:      DefaultFormatConfig(),
			want:     "1.234,56 €",
		},
		{
			name:     "JPY ja-JP",
			locale:   "ja-JP",
			amount:   1234,
			currency: "JPY",
			cfg:      DefaultFormatConfig(),
			want:     "¥1,234",
		},
		{
			name:     "currency code display",
			locale:   "en-US",
			amount:   1234.56,
			currency: "USD",
			cfg:      FormatConfig{CurrencyDisplay: CurrencyCode},
			want:     "USD1,234.56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCurrency(tt.locale, tt.amount, tt.currency, tt.cfg)
			if got != tt.want {
				t.Errorf("FormatCurrency() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		locale string
		time   time.Time
		style  DateStyle
		want   string
	}{
		{
			name:   "en-US short",
			locale: "en-US",
			time:   testTime,
			style:  DateStyleShort,
			want:   "1/15/24",
		},
		{
			name:   "en-US medium",
			locale: "en-US",
			time:   testTime,
			style:  DateStyleMedium,
			want:   "Jan 15, 2024",
		},
		{
			name:   "en-US long",
			locale: "en-US",
			time:   testTime,
			style:  DateStyleLong,
			want:   "January 15, 2024",
		},
		{
			name:   "de-DE short",
			locale: "de-DE",
			time:   testTime,
			style:  DateStyleShort,
			want:   "15.01.24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.locale, tt.time, tt.style)
			if got != tt.want {
				t.Errorf("FormatDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		locale string
		time   time.Time
		want   string
	}{
		{
			name:   "just now",
			locale: "en",
			time:   now.Add(-5 * time.Second),
			want:   "just now",
		},
		{
			name:   "1 minute ago",
			locale: "en",
			time:   now.Add(-1 * time.Minute),
			want:   "1 minute ago",
		},
		{
			name:   "2 minutes ago",
			locale: "en",
			time:   now.Add(-2 * time.Minute),
			want:   "2 minutes ago",
		},
		{
			name:   "1 hour ago",
			locale: "en",
			time:   now.Add(-1 * time.Hour),
			want:   "1 hour ago",
		},
		{
			name:   "yesterday",
			locale: "en",
			time:   now.Add(-24 * time.Hour),
			want:   "yesterday",
		},
		{
			name:   "in 2 hours",
			locale: "en",
			time:   now.Add(2 * time.Hour),
			want:   "in 2 hours",
		},
		{
			name:   "tomorrow",
			locale: "en",
			time:   now.Add(24 * time.Hour),
			want:   "tomorrow",
		},
		{
			name:   "spanish 2 hours ago",
			locale: "es",
			time:   now.Add(-2 * time.Hour),
			want:   "hace 2 horas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTimeFrom(tt.locale, tt.time, now)
			if got != tt.want {
				t.Errorf("FormatRelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		items  []string
		style  ListStyle
		want   string
	}{
		{
			name:   "empty list",
			locale: "en",
			items:  []string{},
			style:  ListStyleAnd,
			want:   "",
		},
		{
			name:   "single item",
			locale: "en",
			items:  []string{"apple"},
			style:  ListStyleAnd,
			want:   "apple",
		},
		{
			name:   "two items",
			locale: "en",
			items:  []string{"apple", "banana"},
			style:  ListStyleAnd,
			want:   "apple and banana",
		},
		{
			name:   "three items and",
			locale: "en",
			items:  []string{"apple", "banana", "cherry"},
			style:  ListStyleAnd,
			want:   "apple, banana, and cherry",
		},
		{
			name:   "three items or",
			locale: "en",
			items:  []string{"apple", "banana", "cherry"},
			style:  ListStyleOr,
			want:   "apple, banana, or cherry",
		},
		{
			name:   "three items narrow",
			locale: "en",
			items:  []string{"apple", "banana", "cherry"},
			style:  ListStyleNarrow,
			want:   "apple, banana, cherry",
		},
		{
			name:   "spanish and",
			locale: "es",
			items:  []string{"manzana", "plátano", "cereza"},
			style:  ListStyleAnd,
			want:   "manzana, plátano y cereza",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatList(tt.locale, tt.items, tt.style)
			if got != tt.want {
				t.Errorf("FormatList() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		number float64
		cfg    FormatConfig
		want   string
	}{
		{
			name:   "basic percent",
			locale: "en-US",
			number: 0.1234,
			cfg:    FormatConfig{MaxDecimals: 2},
			want:   "12.34%",
		},
		{
			name:   "whole number",
			locale: "en-US",
			number: 0.5,
			cfg:    FormatConfig{MaxDecimals: 0},
			want:   "50%",
		},
		{
			name:   "german percent",
			locale: "de-DE",
			number: 0.1234,
			cfg:    FormatConfig{MaxDecimals: 2},
			want:   "12,34%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPercent(tt.locale, tt.number, tt.cfg)
			if got != tt.want {
				t.Errorf("FormatPercent() = %q, want %q", got, tt.want)
			}
		})
	}
}
