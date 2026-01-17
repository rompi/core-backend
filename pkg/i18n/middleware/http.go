// Package middleware provides HTTP and gRPC middleware for i18n.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is used for context values.
type contextKey string

const (
	localeContextKey contextKey = "i18n_locale"
)

// I18n defines the interface required by the middleware.
type I18n interface {
	WithLocale(ctx context.Context, locale string) context.Context
	Locales() []string
}

// HTTPOption configures the HTTP middleware.
type HTTPOption func(*httpMiddleware)

// httpMiddleware extracts locale from HTTP requests.
type httpMiddleware struct {
	i18n            I18n
	queryParam      string
	cookieName      string
	headerName      string
	useAcceptLang   bool
	defaultLocale   string
	localeMatcher   *localeMatcher
	setCookie       bool
	cookieMaxAge    int
	cookiePath      string
	cookieSecure    bool
	cookieHTTPOnly  bool
	cookieSameSite  http.SameSite
}

// localeMatcher matches requested locales to available ones.
type localeMatcher struct {
	available map[string]bool
	languages map[string]string // language -> first matching locale
}

// newLocaleMatcher creates a new locale matcher.
func newLocaleMatcher(locales []string) *localeMatcher {
	m := &localeMatcher{
		available: make(map[string]bool),
		languages: make(map[string]string),
	}

	for _, loc := range locales {
		m.available[loc] = true
		m.available[strings.ToLower(loc)] = true

		// Extract language part
		lang := loc
		if idx := strings.Index(loc, "-"); idx != -1 {
			lang = loc[:idx]
		}
		lang = strings.ToLower(lang)

		// Store first match for language
		if _, ok := m.languages[lang]; !ok {
			m.languages[lang] = loc
		}
	}

	return m
}

// Match finds the best matching locale.
func (m *localeMatcher) Match(requested string) string {
	if requested == "" {
		return ""
	}

	// Normalize
	requested = strings.ReplaceAll(requested, "_", "-")

	// Try exact match
	if m.available[requested] || m.available[strings.ToLower(requested)] {
		return requested
	}

	// Try language only
	lang := requested
	if idx := strings.Index(requested, "-"); idx != -1 {
		lang = requested[:idx]
	}
	lang = strings.ToLower(lang)

	if loc, ok := m.languages[lang]; ok {
		return loc
	}

	return ""
}

// HTTP creates a new HTTP middleware for locale detection.
func HTTP(i18n I18n, opts ...HTTPOption) func(http.Handler) http.Handler {
	m := &httpMiddleware{
		i18n:           i18n,
		defaultLocale:  "en",
		localeMatcher:  newLocaleMatcher(i18n.Locales()),
		cookiePath:     "/",
		cookieMaxAge:   86400 * 365, // 1 year
		cookieSameSite: http.SameSiteLaxMode,
	}

	for _, opt := range opts {
		opt(m)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			locale := m.detectLocale(r)

			// Set cookie if configured
			if m.setCookie && locale != "" && m.cookieName != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     m.cookieName,
					Value:    locale,
					Path:     m.cookiePath,
					MaxAge:   m.cookieMaxAge,
					Secure:   m.cookieSecure,
					HttpOnly: m.cookieHTTPOnly,
					SameSite: m.cookieSameSite,
				})
			}

			// Add locale to context
			ctx := m.i18n.WithLocale(r.Context(), locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// detectLocale detects the locale from the request.
func (m *httpMiddleware) detectLocale(r *http.Request) string {
	var locale string

	// 1. Try query parameter
	if m.queryParam != "" {
		if q := r.URL.Query().Get(m.queryParam); q != "" {
			if matched := m.localeMatcher.Match(q); matched != "" {
				return matched
			}
		}
	}

	// 2. Try cookie
	if m.cookieName != "" {
		if cookie, err := r.Cookie(m.cookieName); err == nil && cookie.Value != "" {
			if matched := m.localeMatcher.Match(cookie.Value); matched != "" {
				return matched
			}
		}
	}

	// 3. Try custom header
	if m.headerName != "" {
		if h := r.Header.Get(m.headerName); h != "" {
			if matched := m.localeMatcher.Match(h); matched != "" {
				return matched
			}
		}
	}

	// 4. Try Accept-Language header
	if m.useAcceptLang {
		locale = m.parseAcceptLanguage(r.Header.Get("Accept-Language"))
		if locale != "" {
			return locale
		}
	}

	// 5. Fall back to default
	return m.defaultLocale
}

// parseAcceptLanguage parses the Accept-Language header and returns the best match.
func (m *httpMiddleware) parseAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}

	// Parse Accept-Language header
	langs := parseAcceptLanguageHeader(header)

	// Find best match
	for _, lang := range langs {
		if matched := m.localeMatcher.Match(lang); matched != "" {
			return matched
		}
	}

	return ""
}

// parseAcceptLanguageHeader parses an Accept-Language header and returns locales in order of preference.
func parseAcceptLanguageHeader(header string) []string {
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
				qStr := qPart[2:]
				if q := parseFloat(qStr); q >= 0 {
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

// parseFloat parses a float from a string without importing strconv.
func parseFloat(s string) float64 {
	if s == "" {
		return -1
	}

	var result float64
	var decimalSeen bool
	var decimalPlace float64 = 0.1

	for _, c := range s {
		if c == '.' {
			if decimalSeen {
				return -1
			}
			decimalSeen = true
			continue
		}

		if c < '0' || c > '9' {
			return -1
		}

		digit := float64(c - '0')
		if decimalSeen {
			result += digit * decimalPlace
			decimalPlace *= 0.1
		} else {
			result = result*10 + digit
		}
	}

	return result
}

// WithQueryParam configures the query parameter to check for locale.
func WithQueryParam(param string) HTTPOption {
	return func(m *httpMiddleware) {
		m.queryParam = param
	}
}

// WithCookie configures the cookie name to check for locale.
func WithCookie(name string) HTTPOption {
	return func(m *httpMiddleware) {
		m.cookieName = name
	}
}

// WithHeader configures a custom header to check for locale.
func WithHeader(name string) HTTPOption {
	return func(m *httpMiddleware) {
		m.headerName = name
	}
}

// WithAcceptLanguage enables parsing of the Accept-Language header.
func WithAcceptLanguage() HTTPOption {
	return func(m *httpMiddleware) {
		m.useAcceptLang = true
	}
}

// WithDefaultLocale sets the default locale when none is detected.
func WithDefaultLocale(locale string) HTTPOption {
	return func(m *httpMiddleware) {
		m.defaultLocale = locale
	}
}

// WithSetCookie enables setting a cookie with the detected locale.
func WithSetCookie(enabled bool) HTTPOption {
	return func(m *httpMiddleware) {
		m.setCookie = enabled
	}
}

// WithCookieConfig configures the locale cookie settings.
func WithCookieConfig(maxAge int, path string, secure, httpOnly bool, sameSite http.SameSite) HTTPOption {
	return func(m *httpMiddleware) {
		m.cookieMaxAge = maxAge
		m.cookiePath = path
		m.cookieSecure = secure
		m.cookieHTTPOnly = httpOnly
		m.cookieSameSite = sameSite
	}
}

// LocaleFromContext extracts the locale from a context.
func LocaleFromContext(ctx context.Context) string {
	if v := ctx.Value(localeContextKey); v != nil {
		if locale, ok := v.(string); ok {
			return locale
		}
	}
	return ""
}

// ContextWithLocale returns a new context with the specified locale.
func ContextWithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey, locale)
}
