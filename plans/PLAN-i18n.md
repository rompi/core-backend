# Package Plan: pkg/i18n

## Overview

An internationalization (i18n) and localization (l10n) package for building multi-language applications. Provides message translation, pluralization, number/date/currency formatting, and locale management with support for multiple translation backends.

## Goals

1. **Message Translation** - Key-based message lookup with interpolation
2. **Pluralization** - Language-aware plural forms
3. **Formatting** - Numbers, dates, currencies, relative time
4. **Multiple Backends** - JSON, YAML, PO/MO files, database
5. **Fallback Chain** - Locale fallback (en-US → en → default)
6. **Hot Reload** - Update translations without restart
7. **Context Propagation** - Locale via context
8. **Zero Dependencies** - Core uses stdlib only

## Architecture

```
pkg/i18n/
├── i18n.go               # Core I18n interface
├── config.go             # Configuration
├── options.go            # Functional options
├── locale.go             # Locale parsing and matching
├── message.go            # Message definition
├── plural.go             # Pluralization rules
├── format/
│   ├── number.go         # Number formatting
│   ├── currency.go       # Currency formatting
│   ├── datetime.go       # Date/time formatting
│   ├── relative.go       # Relative time (2 hours ago)
│   └── list.go           # List formatting (a, b, and c)
├── catalog/
│   ├── catalog.go        # Message catalog interface
│   ├── json.go           # JSON file catalog
│   ├── yaml.go           # YAML file catalog
│   ├── po.go             # Gettext PO/MO files
│   ├── database.go       # Database catalog
│   └── embed.go          # Embedded FS catalog
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── extract/
│   ├── extract.go        # Message extraction
│   └── ast.go            # Go AST parsing
├── examples/
│   ├── basic/
│   ├── http-server/
│   ├── pluralization/
│   └── formatting/
└── README.md
```

## Core Interfaces

```go
package i18n

import (
    "context"
    "time"
)

// I18n provides internationalization functionality
type I18n interface {
    // T translates a message key
    T(ctx context.Context, key string, args ...interface{}) string

    // Tn translates with pluralization
    Tn(ctx context.Context, key string, count int, args ...interface{}) string

    // Tf translates with named arguments
    Tf(ctx context.Context, key string, args map[string]interface{}) string

    // L returns a locale-specific localizer
    L(locale string) Localizer

    // Locale returns the locale from context
    Locale(ctx context.Context) string

    // WithLocale returns context with locale
    WithLocale(ctx context.Context, locale string) context.Context

    // Locales returns available locales
    Locales() []string

    // Reload reloads translations
    Reload() error
}

// Localizer provides locale-specific operations
type Localizer interface {
    // Translation
    T(key string, args ...interface{}) string
    Tn(key string, count int, args ...interface{}) string
    Tf(key string, args map[string]interface{}) string

    // Formatting
    FormatNumber(n float64, opts ...FormatOption) string
    FormatCurrency(amount float64, currency string, opts ...FormatOption) string
    FormatDate(t time.Time, style DateStyle) string
    FormatTime(t time.Time, style TimeStyle) string
    FormatDateTime(t time.Time, dateStyle DateStyle, timeStyle TimeStyle) string
    FormatRelativeTime(t time.Time) string
    FormatList(items []string, style ListStyle) string
    FormatPercent(n float64, opts ...FormatOption) string

    // Locale info
    Locale() string
    Language() string
    Region() string
    Direction() Direction // LTR or RTL
}

// Message represents a translatable message
type Message struct {
    // ID is the message identifier
    ID string

    // Description for translators
    Description string

    // One is the singular form
    One string

    // Other is the plural form
    Other string

    // Zero, Two, Few, Many for complex pluralization
    Zero string
    Two  string
    Few  string
    Many string
}

// Catalog provides message storage
type Catalog interface {
    // Lookup finds a message
    Lookup(locale, key string) (*Message, error)

    // All returns all messages for a locale
    All(locale string) (map[string]*Message, error)

    // Locales returns available locales
    Locales() []string

    // Reload reloads the catalog
    Reload() error
}

// Direction for text direction
type Direction string

const (
    LTR Direction = "ltr"
    RTL Direction = "rtl"
)

// DateStyle for date formatting
type DateStyle int

const (
    DateStyleShort  DateStyle = iota // 1/15/24
    DateStyleMedium                   // Jan 15, 2024
    DateStyleLong                     // January 15, 2024
    DateStyleFull                     // Monday, January 15, 2024
)

// TimeStyle for time formatting
type TimeStyle int

const (
    TimeStyleShort TimeStyle = iota // 3:04 PM
    TimeStyleMedium                  // 3:04:05 PM
    TimeStyleLong                    // 3:04:05 PM EST
)

// ListStyle for list formatting
type ListStyle int

const (
    ListStyleAnd         ListStyle = iota // a, b, and c
    ListStyleOr                           // a, b, or c
    ListStyleNarrow                       // a, b, c
)
```

## Configuration

```go
// Config holds i18n configuration
type Config struct {
    // Default locale
    DefaultLocale string `env:"I18N_DEFAULT_LOCALE" default:"en"`

    // Fallback locale (when translation missing)
    FallbackLocale string `env:"I18N_FALLBACK_LOCALE" default:"en"`

    // Catalog type: "json", "yaml", "po", "database", "embed"
    CatalogType string `env:"I18N_CATALOG_TYPE" default:"json"`

    // Path to translation files
    Path string `env:"I18N_PATH" default:"./locales"`

    // Enable hot reload
    HotReload bool `env:"I18N_HOT_RELOAD" default:"false"`

    // Missing key behavior: "key", "empty", "error"
    MissingKeyBehavior string `env:"I18N_MISSING_KEY" default:"key"`

    // Log missing translations
    LogMissing bool `env:"I18N_LOG_MISSING" default:"true"`
}

// FormatOption configures formatting
type FormatOption func(*formatConfig)

// WithMinDecimals sets minimum decimal places
func WithMinDecimals(n int) FormatOption

// WithMaxDecimals sets maximum decimal places
func WithMaxDecimals(n int) FormatOption

// WithGrouping enables/disables thousand separators
func WithGrouping(enabled bool) FormatOption

// WithCurrencyDisplay sets currency display mode
func WithCurrencyDisplay(display CurrencyDisplay) FormatOption

type CurrencyDisplay int

const (
    CurrencySymbol CurrencyDisplay = iota // $
    CurrencyCode                          // USD
    CurrencyName                          // US Dollar
)
```

## Translation File Formats

### JSON Format

```json
// locales/en.json
{
  "greeting": "Hello, {{.Name}}!",
  "items": {
    "one": "{{.Count}} item",
    "other": "{{.Count}} items"
  },
  "welcome": {
    "title": "Welcome",
    "message": "Welcome to our application"
  },
  "errors": {
    "not_found": "Resource not found",
    "unauthorized": "You are not authorized"
  }
}

// locales/es.json
{
  "greeting": "¡Hola, {{.Name}}!",
  "items": {
    "one": "{{.Count}} artículo",
    "other": "{{.Count}} artículos"
  },
  "welcome": {
    "title": "Bienvenido",
    "message": "Bienvenido a nuestra aplicación"
  }
}
```

### YAML Format

```yaml
# locales/en.yaml
greeting: "Hello, {{.Name}}!"

items:
  one: "{{.Count}} item"
  other: "{{.Count}} items"

welcome:
  title: "Welcome"
  message: "Welcome to our application"

errors:
  not_found: "Resource not found"
  unauthorized: "You are not authorized"
```

### Gettext PO Format

```po
# locales/es/LC_MESSAGES/messages.po
msgid "greeting"
msgstr "¡Hola, %s!"

msgid "item"
msgid_plural "items"
msgstr[0] "%d artículo"
msgstr[1] "%d artículos"
```

## Pluralization Rules

```go
// Plural categories per CLDR
type PluralCategory int

const (
    Zero PluralCategory = iota
    One
    Two
    Few
    Many
    Other
)

// Built-in rules for common languages
// English: one (n=1), other
// French: one (n=0,1), other
// Russian: one (n%10=1 && n%100!=11), few (n%10=2..4 && n%100!=12..14), many, other
// Arabic: zero (n=0), one (n=1), two (n=2), few (n%100=3..10), many (n%100=11..99), other
// Japanese/Chinese/Korean: other (no plural)

// RegisterPluralRule registers a custom plural rule
func RegisterPluralRule(locale string, rule PluralRule)

type PluralRule func(n int) PluralCategory
```

## Usage Examples

### Basic Translation

```go
package main

import (
    "context"
    "fmt"
    "github.com/user/core-backend/pkg/i18n"
)

func main() {
    // Create i18n instance
    i, err := i18n.New(i18n.Config{
        DefaultLocale: "en",
        Path:          "./locales",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Default locale (English)
    fmt.Println(i.T(ctx, "greeting", "World"))
    // Output: Hello, World!

    // Spanish
    ctx = i.WithLocale(ctx, "es")
    fmt.Println(i.T(ctx, "greeting", "Mundo"))
    // Output: ¡Hola, Mundo!
}
```

### Named Arguments

```go
func main() {
    i, _ := i18n.New(cfg)
    ctx := context.Background()

    // Using named arguments
    msg := i.Tf(ctx, "order.confirmation", map[string]interface{}{
        "OrderID":  "12345",
        "Name":     "John",
        "Total":    99.99,
        "Currency": "USD",
    })
    // Template: "Order #{{.OrderID}} confirmed. Thank you, {{.Name}}!"
    // Output: "Order #12345 confirmed. Thank you, John!"
}
```

### Pluralization

```go
func main() {
    i, _ := i18n.New(cfg)
    ctx := context.Background()

    // English
    fmt.Println(i.Tn(ctx, "items", 0))  // 0 items
    fmt.Println(i.Tn(ctx, "items", 1))  // 1 item
    fmt.Println(i.Tn(ctx, "items", 5))  // 5 items

    // Russian (complex pluralization)
    ctx = i.WithLocale(ctx, "ru")
    fmt.Println(i.Tn(ctx, "items", 1))   // 1 товар
    fmt.Println(i.Tn(ctx, "items", 2))   // 2 товара
    fmt.Println(i.Tn(ctx, "items", 5))   // 5 товаров
    fmt.Println(i.Tn(ctx, "items", 21))  // 21 товар
}
```

### Number Formatting

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en-US")
    de := i.L("de-DE")
    fr := i.L("fr-FR")

    n := 1234567.89

    fmt.Println(en.FormatNumber(n)) // 1,234,567.89
    fmt.Println(de.FormatNumber(n)) // 1.234.567,89
    fmt.Println(fr.FormatNumber(n)) // 1 234 567,89

    // With options
    fmt.Println(en.FormatNumber(n,
        i18n.WithMinDecimals(2),
        i18n.WithMaxDecimals(2),
    ))
}
```

### Currency Formatting

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en-US")
    de := i.L("de-DE")
    ja := i.L("ja-JP")

    amount := 1234.56

    // US Dollar
    fmt.Println(en.FormatCurrency(amount, "USD")) // $1,234.56
    fmt.Println(de.FormatCurrency(amount, "USD")) // 1.234,56 $
    fmt.Println(de.FormatCurrency(amount, "EUR")) // 1.234,56 €

    // Japanese Yen (no decimals)
    fmt.Println(ja.FormatCurrency(1234, "JPY")) // ¥1,234

    // Currency code display
    fmt.Println(en.FormatCurrency(amount, "USD",
        i18n.WithCurrencyDisplay(i18n.CurrencyCode),
    )) // USD 1,234.56
}
```

### Date/Time Formatting

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en-US")
    de := i.L("de-DE")
    ja := i.L("ja-JP")

    t := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

    // Date formatting
    fmt.Println(en.FormatDate(t, i18n.DateStyleShort))  // 1/15/24
    fmt.Println(en.FormatDate(t, i18n.DateStyleMedium)) // Jan 15, 2024
    fmt.Println(en.FormatDate(t, i18n.DateStyleLong))   // January 15, 2024
    fmt.Println(en.FormatDate(t, i18n.DateStyleFull))   // Monday, January 15, 2024

    fmt.Println(de.FormatDate(t, i18n.DateStyleMedium)) // 15. Jan. 2024
    fmt.Println(ja.FormatDate(t, i18n.DateStyleMedium)) // 2024年1月15日

    // Time formatting
    fmt.Println(en.FormatTime(t, i18n.TimeStyleShort))  // 2:30 PM
    fmt.Println(de.FormatTime(t, i18n.TimeStyleShort))  // 14:30

    // Combined
    fmt.Println(en.FormatDateTime(t, i18n.DateStyleMedium, i18n.TimeStyleShort))
    // Jan 15, 2024, 2:30 PM
}
```

### Relative Time

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en-US")
    es := i.L("es")

    now := time.Now()

    fmt.Println(en.FormatRelativeTime(now.Add(-5 * time.Second)))  // just now
    fmt.Println(en.FormatRelativeTime(now.Add(-2 * time.Minute)))  // 2 minutes ago
    fmt.Println(en.FormatRelativeTime(now.Add(-3 * time.Hour)))    // 3 hours ago
    fmt.Println(en.FormatRelativeTime(now.Add(-1 * 24 * time.Hour))) // yesterday
    fmt.Println(en.FormatRelativeTime(now.Add(-7 * 24 * time.Hour))) // last week
    fmt.Println(en.FormatRelativeTime(now.Add(2 * time.Hour)))     // in 2 hours

    fmt.Println(es.FormatRelativeTime(now.Add(-2 * time.Hour)))    // hace 2 horas
}
```

### List Formatting

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en-US")
    es := i.L("es")

    items := []string{"apples", "oranges", "bananas"}

    fmt.Println(en.FormatList(items, i18n.ListStyleAnd))
    // apples, oranges, and bananas

    fmt.Println(en.FormatList(items, i18n.ListStyleOr))
    // apples, oranges, or bananas

    fmt.Println(es.FormatList(items, i18n.ListStyleAnd))
    // apples, oranges y bananas
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/i18n"
    "github.com/user/core-backend/pkg/i18n/middleware"
)

func main() {
    i, _ := i18n.New(cfg)

    // Middleware extracts locale from:
    // 1. Query param: ?lang=es
    // 2. Cookie: lang=es
    // 3. Accept-Language header
    mw := middleware.HTTP(i,
        middleware.WithQueryParam("lang"),
        middleware.WithCookie("lang"),
        middleware.WithAcceptLanguage(),
    )

    mux := http.NewServeMux()
    http.ListenAndServe(":8080", mw(mux))
}

// In handlers, use context
func handler(w http.ResponseWriter, r *http.Request) {
    msg := i18n.T(r.Context(), "welcome.message")
    fmt.Fprintf(w, msg)
}
```

### Embedded Translations

```go
import (
    "embed"
    "github.com/user/core-backend/pkg/i18n"
    "github.com/user/core-backend/pkg/i18n/catalog"
)

//go:embed locales/*.json
var localesFS embed.FS

func main() {
    cat, _ := catalog.NewEmbed(localesFS, "locales")

    i, _ := i18n.New(i18n.Config{
        DefaultLocale: "en",
    }, i18n.WithCatalog(cat))

    // Translations are embedded in binary
}
```

### Database Catalog

```go
import (
    "github.com/user/core-backend/pkg/i18n/catalog"
)

func main() {
    // Store translations in database
    dbCatalog, _ := catalog.NewDatabase(db, catalog.DatabaseConfig{
        TableName: "translations",
    })

    i, _ := i18n.New(cfg, i18n.WithCatalog(dbCatalog))

    // Translations can be managed via admin UI
    // Hot reload on changes
}

// Database schema
/*
CREATE TABLE translations (
    id SERIAL PRIMARY KEY,
    locale VARCHAR(10) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    plural_one TEXT,
    plural_other TEXT,
    plural_few TEXT,
    plural_many TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(locale, key)
);
*/
```

### Message Extraction

```go
// Extract translatable strings from Go code
// go run ./cmd/i18n-extract ./...

// Scans for:
// i.T(ctx, "key", ...)
// i.Tn(ctx, "key", ...)
// i.Tf(ctx, "key", ...)

// Outputs to JSON/PO file with new keys
```

### RTL Support

```go
func main() {
    i, _ := i18n.New(cfg)

    en := i.L("en")
    ar := i.L("ar") // Arabic
    he := i.L("he") // Hebrew

    fmt.Println(en.Direction()) // ltr
    fmt.Println(ar.Direction()) // rtl
    fmt.Println(he.Direction()) // rtl

    // Use in templates
    // <html dir="{{.Localizer.Direction}}">
}
```

### Nested Keys

```go
// locales/en.json
{
    "errors": {
        "validation": {
            "required": "This field is required",
            "email": "Invalid email address",
            "min_length": "Must be at least {{.Min}} characters"
        }
    }
}

// Usage
i.T(ctx, "errors.validation.required")
i.Tf(ctx, "errors.validation.min_length", map[string]interface{}{
    "Min": 8,
})
```

### Missing Translation Handling

```go
func main() {
    i, _ := i18n.New(i18n.Config{
        DefaultLocale:      "en",
        FallbackLocale:     "en",
        MissingKeyBehavior: "key",    // Return the key itself
        LogMissing:         true,      // Log missing translations
    },
        i18n.WithMissingHandler(func(locale, key string) {
            log.Printf("Missing translation: %s/%s", locale, key)
            metrics.IncrCounter("i18n.missing", 1)
        }),
    )
}
```

## Error Handling

```go
var (
    // ErrLocaleNotFound is returned for unknown locale
    ErrLocaleNotFound = errors.New("i18n: locale not found")

    // ErrKeyNotFound is returned for unknown translation key
    ErrKeyNotFound = errors.New("i18n: key not found")

    // ErrInvalidFormat is returned for invalid translation format
    ErrInvalidFormat = errors.New("i18n: invalid format")

    // ErrCatalogLoad is returned when catalog fails to load
    ErrCatalogLoad = errors.New("i18n: catalog load failed")
)
```

## Dependencies

- **Required:** None (core uses stdlib)
- **Optional:**
  - `gopkg.in/yaml.v3` for YAML catalog
  - `github.com/leonelquinteros/gotext` for PO/MO files
  - `golang.org/x/text` for advanced formatting

## Test Coverage Requirements

- Unit tests for all public functions
- Pluralization tests for major languages
- Formatting tests across locales
- Fallback chain tests
- Hot reload tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & JSON Catalog
1. Define I18n, Localizer interfaces
2. JSON file catalog
3. Basic translation with interpolation
4. Context-based locale

### Phase 2: Pluralization
1. CLDR plural rules
2. Common language rules (en, es, fr, de, ru, ar, zh, ja)
3. Custom rule registration

### Phase 3: Formatting
1. Number formatting
2. Currency formatting
3. Date/time formatting
4. Relative time

### Phase 4: Additional Catalogs
1. YAML catalog
2. PO/MO catalog (Gettext)
3. Database catalog
4. Embedded FS catalog

### Phase 5: Middleware & Integration
1. HTTP middleware
2. gRPC interceptor
3. Template helpers

### Phase 6: Tooling
1. Message extraction tool
2. Missing translation detection
3. Translation file validation

### Phase 7: Documentation & Examples
1. README with full documentation
2. Basic example
3. Pluralization example
4. Formatting example
5. HTTP server example
