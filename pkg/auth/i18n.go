package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var englishMessages = map[string]string{
	"invalid_credentials": "Invalid email or password",
	"user_already_exists": "A user with that email already exists",
	"user_not_found":      "User not found",
	"account_locked":      "Account is locked due to too many failed attempts",
	"invalid_token":       "Token is invalid or expired",
	"weak_password":       "Password does not meet complexity requirements",
	"rate_limit_exceeded": "Too many requests, please try again later",
	"permission_denied":   "You do not have permission to perform this action",
	"session_expired":     "Session has expired",
	"invalid_reset_token": "Reset token is invalid or expired",
}

// DefaultTranslator is the shared translator used by auth errors and handlers.
var DefaultTranslator = NewTranslator("en")

// Translator resolves localized messages for auth error codes.
type Translator struct {
	defaultLanguage string
	mu              sync.RWMutex
	messages        map[string]map[string]string
}

// NewTranslator creates a Translator with the supplied default language.
func NewTranslator(defaultLanguage string) *Translator {
	lang := normalizeLanguage(defaultLanguage)
	if lang == "" {
		lang = "en"
	}
	t := &Translator{
		defaultLanguage: lang,
		messages:        make(map[string]map[string]string),
	}
	t.Register("en", englishMessages)
	return t
}

// Register stores translation messages for a language.
func (t *Translator) Register(language string, entries map[string]string) {
	lang := normalizeLanguage(language)
	if lang == "" || len(entries) == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	merged := make(map[string]string)
	if current, ok := t.messages[lang]; ok {
		for key, value := range current {
			merged[key] = value
		}
	}
	for key, value := range entries {
		merged[key] = value
	}
	t.messages[lang] = merged
}

// LoadFromFile adds translations by reading the JSON file at path.
func (t *Translator) LoadFromFile(language, path string) error {
	clean, err := safeTranslationPath(path)
	if err != nil {
		return err
	}
	// #nosec G304 -- path is validated in safeTranslationPath.
	data, err := os.ReadFile(clean)
	if err != nil {
		return fmt.Errorf("read translations: %w", err)
	}
	return t.LoadFromBytes(language, data)
}

// LoadFromBytes adds translations from JSON data.
func (t *Translator) LoadFromBytes(language string, data []byte) error {
	entries := map[string]string{}
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("parse translations: %w", err)
	}
	t.Register(language, entries)
	return nil
}

// Message resolves the localized message for code in the requested language.
// If the code does not exist, it falls back to the default language or the code itself.
func (t *Translator) Message(language, code string, args ...interface{}) string {
	if code == "" {
		return ""
	}
	lang := normalizeLanguage(language)
	if lang == "" {
		lang = t.defaultLanguage
	}
	if msg, found := t.lookup(lang, code); found {
		return formatMessage(msg, args...)
	}
	if msg, found := t.lookup(t.defaultLanguage, code); found {
		return formatMessage(msg, args...)
	}
	if len(args) > 0 {
		return formatMessage(code, args...)
	}
	return code
}

// Languages returns the registered languages.
func (t *Translator) Languages() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	langs := make([]string, 0, len(t.messages))
	for lang := range t.messages {
		langs = append(langs, lang)
	}
	return langs
}

func (t *Translator) lookup(language, code string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if entries, ok := t.messages[language]; ok {
		msg, ok := entries[code]
		return msg, ok
	}
	return "", false
}

func normalizeLanguage(language string) string {
	return strings.ToLower(strings.TrimSpace(language))
}

func safeTranslationPath(path string) (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determine working directory: %w", err)
	}
	abs := path
	if !filepath.IsAbs(path) {
		abs = filepath.Join(root, path)
	}
	clean := filepath.Clean(abs)
	rootWithSep := root
	if !strings.HasSuffix(rootWithSep, string(os.PathSeparator)) {
		rootWithSep += string(os.PathSeparator)
	}
	if clean != root && !strings.HasPrefix(clean, rootWithSep) {
		return "", fmt.Errorf("translation path %s outside working directory", path)
	}
	return clean, nil
}

func formatMessage(template string, args ...interface{}) string {
	if template == "" || len(args) == 0 {
		return template
	}
	result := template
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(arg))
	}
	return result
}
