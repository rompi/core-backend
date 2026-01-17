package i18n

import (
	"fmt"
	"strings"
)

// Direction represents text direction.
type Direction string

const (
	// LTR represents left-to-right text direction.
	LTR Direction = "ltr"

	// RTL represents right-to-left text direction.
	RTL Direction = "rtl"
)

// Locale represents a parsed locale identifier.
type Locale struct {
	// Language is the ISO 639-1 language code (e.g., "en", "es").
	Language string

	// Region is the ISO 3166-1 alpha-2 region code (e.g., "US", "GB").
	Region string

	// Script is the ISO 15924 script code (e.g., "Hans", "Hant").
	Script string
}

// ParseLocale parses a locale string into a Locale struct.
// Supported formats: "en", "en-US", "en_US", "zh-Hans", "zh-Hans-CN".
func ParseLocale(s string) (*Locale, error) {
	if s == "" {
		return nil, fmt.Errorf("%w: empty locale string", ErrInvalidLocale)
	}

	// Normalize separator to hyphen
	s = strings.ReplaceAll(s, "_", "-")
	parts := strings.Split(s, "-")

	locale := &Locale{}

	// First part is always the language
	locale.Language = strings.ToLower(parts[0])
	if len(locale.Language) < 2 || len(locale.Language) > 3 {
		return nil, fmt.Errorf("%w: invalid language code: %s", ErrInvalidLocale, parts[0])
	}

	if len(parts) == 1 {
		return locale, nil
	}

	// Handle second part (can be script or region)
	if len(parts) >= 2 {
		part := parts[1]
		if len(part) == 4 {
			// 4-letter code is a script (e.g., "Hans", "Hant")
			locale.Script = strings.Title(strings.ToLower(part))
		} else if len(part) == 2 {
			// 2-letter code is a region
			locale.Region = strings.ToUpper(part)
		} else {
			return nil, fmt.Errorf("%w: invalid locale part: %s", ErrInvalidLocale, part)
		}
	}

	// Handle third part (must be region if script was second)
	if len(parts) >= 3 {
		part := parts[2]
		if len(part) == 2 {
			locale.Region = strings.ToUpper(part)
		} else {
			return nil, fmt.Errorf("%w: invalid region code: %s", ErrInvalidLocale, part)
		}
	}

	return locale, nil
}

// String returns the canonical string representation of the locale.
func (l *Locale) String() string {
	if l.Script != "" && l.Region != "" {
		return fmt.Sprintf("%s-%s-%s", l.Language, l.Script, l.Region)
	}
	if l.Script != "" {
		return fmt.Sprintf("%s-%s", l.Language, l.Script)
	}
	if l.Region != "" {
		return fmt.Sprintf("%s-%s", l.Language, l.Region)
	}
	return l.Language
}

// Direction returns the text direction for the locale.
func (l *Locale) Direction() Direction {
	return GetLanguageDirection(l.Language)
}

// GetLanguageDirection returns the text direction for a language code.
func GetLanguageDirection(lang string) Direction {
	// RTL languages
	rtlLanguages := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian
		"ur": true, // Urdu
		"yi": true, // Yiddish
		"ps": true, // Pashto
		"sd": true, // Sindhi
		"ug": true, // Uyghur
	}

	if rtlLanguages[strings.ToLower(lang)] {
		return RTL
	}
	return LTR
}

// LocaleMatcher finds the best matching locale from available locales.
type LocaleMatcher struct {
	available []string
	parsed    map[string]*Locale
}

// NewLocaleMatcher creates a new locale matcher with the given available locales.
func NewLocaleMatcher(available []string) *LocaleMatcher {
	m := &LocaleMatcher{
		available: available,
		parsed:    make(map[string]*Locale),
	}

	for _, loc := range available {
		if parsed, err := ParseLocale(loc); err == nil {
			m.parsed[loc] = parsed
		}
	}

	return m
}

// Match finds the best matching locale for the requested locale.
// It tries exact match first, then language-region, then language only.
func (m *LocaleMatcher) Match(requested string) string {
	if requested == "" {
		return ""
	}

	reqLocale, err := ParseLocale(requested)
	if err != nil {
		return ""
	}

	// Try exact match
	for _, avail := range m.available {
		if strings.EqualFold(avail, requested) {
			return avail
		}
		if parsed, ok := m.parsed[avail]; ok {
			if strings.EqualFold(parsed.String(), reqLocale.String()) {
				return avail
			}
		}
	}

	// Try language-region match (ignoring script)
	if reqLocale.Region != "" {
		for avail, parsed := range m.parsed {
			if strings.EqualFold(parsed.Language, reqLocale.Language) &&
				strings.EqualFold(parsed.Region, reqLocale.Region) {
				return avail
			}
		}
	}

	// Try language-only match
	for avail, parsed := range m.parsed {
		if strings.EqualFold(parsed.Language, reqLocale.Language) {
			return avail
		}
	}

	return ""
}

// ParseAcceptLanguage parses an Accept-Language header and returns locales in order of preference.
func ParseAcceptLanguage(header string) []string {
	if header == "" {
		return nil
	}

	type langQuality struct {
		lang    string
		quality float64
	}

	var langs []langQuality

	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lang := part
		quality := 1.0

		if idx := strings.Index(part, ";"); idx != -1 {
			lang = strings.TrimSpace(part[:idx])
			qPart := strings.TrimSpace(part[idx+1:])
			if strings.HasPrefix(qPart, "q=") {
				var q float64
				if _, err := fmt.Sscanf(qPart, "q=%f", &q); err == nil {
					quality = q
				}
			}
		}

		if lang != "" && lang != "*" {
			langs = append(langs, langQuality{lang: lang, quality: quality})
		}
	}

	// Sort by quality (descending)
	for i := 0; i < len(langs)-1; i++ {
		for j := i + 1; j < len(langs); j++ {
			if langs[j].quality > langs[i].quality {
				langs[i], langs[j] = langs[j], langs[i]
			}
		}
	}

	result := make([]string, len(langs))
	for i, lq := range langs {
		result[i] = lq.lang
	}

	return result
}
