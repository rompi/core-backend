package i18n

import (
	"reflect"
	"testing"
)

func TestParseLocale(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Locale
		wantErr bool
	}{
		{
			name:  "language only",
			input: "en",
			want:  &Locale{Language: "en"},
		},
		{
			name:  "language and region with hyphen",
			input: "en-US",
			want:  &Locale{Language: "en", Region: "US"},
		},
		{
			name:  "language and region with underscore",
			input: "en_US",
			want:  &Locale{Language: "en", Region: "US"},
		},
		{
			name:  "language and script",
			input: "zh-Hans",
			want:  &Locale{Language: "zh", Script: "Hans"},
		},
		{
			name:  "full locale",
			input: "zh-Hans-CN",
			want:  &Locale{Language: "zh", Script: "Hans", Region: "CN"},
		},
		{
			name:  "case normalization",
			input: "EN-us",
			want:  &Locale{Language: "en", Region: "US"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid language",
			input:   "x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLocale(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLocale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLocale() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLocale_String(t *testing.T) {
	tests := []struct {
		name   string
		locale Locale
		want   string
	}{
		{
			name:   "language only",
			locale: Locale{Language: "en"},
			want:   "en",
		},
		{
			name:   "language and region",
			locale: Locale{Language: "en", Region: "US"},
			want:   "en-US",
		},
		{
			name:   "language and script",
			locale: Locale{Language: "zh", Script: "Hans"},
			want:   "zh-Hans",
		},
		{
			name:   "full locale",
			locale: Locale{Language: "zh", Script: "Hans", Region: "CN"},
			want:   "zh-Hans-CN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.locale.String(); got != tt.want {
				t.Errorf("Locale.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocale_Direction(t *testing.T) {
	tests := []struct {
		name   string
		locale Locale
		want   Direction
	}{
		{
			name:   "English LTR",
			locale: Locale{Language: "en"},
			want:   LTR,
		},
		{
			name:   "Arabic RTL",
			locale: Locale{Language: "ar"},
			want:   RTL,
		},
		{
			name:   "Hebrew RTL",
			locale: Locale{Language: "he"},
			want:   RTL,
		},
		{
			name:   "Persian RTL",
			locale: Locale{Language: "fa"},
			want:   RTL,
		},
		{
			name:   "German LTR",
			locale: Locale{Language: "de"},
			want:   LTR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.locale.Direction(); got != tt.want {
				t.Errorf("Locale.Direction() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocaleMatcher_Match(t *testing.T) {
	matcher := NewLocaleMatcher([]string{"en", "en-US", "es", "fr-FR", "zh-Hans"})

	tests := []struct {
		name      string
		requested string
		want      string
	}{
		{
			name:      "exact match",
			requested: "en-US",
			want:      "en-US",
		},
		{
			name:      "language match",
			requested: "en-GB",
			want:      "en",
		},
		{
			name:      "case insensitive",
			requested: "EN-us",
			want:      "en-US",
		},
		{
			name:      "script match",
			requested: "zh-Hans-CN",
			want:      "zh-Hans",
		},
		{
			name:      "no match",
			requested: "de",
			want:      "",
		},
		{
			name:      "empty string",
			requested: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.Match(tt.requested); got != tt.want {
				t.Errorf("LocaleMatcher.Match() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   []string
	}{
		{
			name:   "single language",
			header: "en",
			want:   []string{"en"},
		},
		{
			name:   "multiple languages",
			header: "en, es, fr",
			want:   []string{"en", "es", "fr"},
		},
		{
			name:   "with quality",
			header: "en-US,en;q=0.9,es;q=0.8",
			want:   []string{"en-US", "en", "es"},
		},
		{
			name:   "sorted by quality",
			header: "es;q=0.5,en;q=0.9,fr;q=0.7",
			want:   []string{"en", "fr", "es"},
		},
		{
			name:   "empty header",
			header: "",
			want:   nil,
		},
		{
			name:   "wildcard ignored",
			header: "en, *;q=0.1",
			want:   []string{"en"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAcceptLanguage(tt.header)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAcceptLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}
