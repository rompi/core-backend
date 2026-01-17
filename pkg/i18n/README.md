# pkg/i18n

An internationalization (i18n) and localization (l10n) package for building multi-language applications. Provides message translation, pluralization, number/date/currency formatting, and locale management with support for multiple translation backends.

## Features

- **Message Translation** - Key-based message lookup with interpolation
- **Pluralization** - Language-aware plural forms (CLDR rules for 30+ languages)
- **Formatting** - Numbers, dates, currencies, relative time, lists, percentages
- **Multiple Backends** - JSON, YAML, embedded filesystem, in-memory
- **Fallback Chain** - Locale fallback (en-US → en → default)
- **Hot Reload** - Update translations without restart
- **Context Propagation** - Locale via Go context
- **HTTP Middleware** - Automatic locale detection from headers, cookies, query params
- **RTL Support** - Text direction detection for Arabic, Hebrew, etc.
- **Zero Core Dependencies** - Core functionality uses stdlib only

## Installation

```go
import "github.com/rompi/core-backend/pkg/i18n"
```

## Quick Start

### Basic Translation

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rompi/core-backend/pkg/i18n"
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

### Translation Files

Create JSON files in your locales directory:

```json
// locales/en.json
{
  "greeting": "Hello, {{.Arg0}}!",
  "items": {
    "one": "{{.Count}} item",
    "other": "{{.Count}} items"
  },
  "welcome": {
    "title": "Welcome",
    "message": "Welcome to our application"
  }
}
```

```json
// locales/es.json
{
  "greeting": "¡Hola, {{.Arg0}}!",
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

## Configuration

| Field | Environment Variable | Default | Description |
|-------|---------------------|---------|-------------|
| DefaultLocale | I18N_DEFAULT_LOCALE | "en" | Default locale when none specified |
| FallbackLocale | I18N_FALLBACK_LOCALE | "en" | Fallback when translation missing |
| CatalogType | I18N_CATALOG_TYPE | "json" | Catalog type: json, yaml |
| Path | I18N_PATH | "./locales" | Path to translation files |
| HotReload | I18N_HOT_RELOAD | false | Enable hot reload |
| MissingKeyBehavior | I18N_MISSING_KEY | "key" | key, empty, or error |
| LogMissing | I18N_LOG_MISSING | true | Log missing translations |

## Translation Methods

### Simple Translation (T)

```go
// Positional arguments
i.T(ctx, "greeting", "World")
// Template: "Hello, {{.Arg0}}!"
// Output: "Hello, World!"
```

### Named Arguments (Tf)

```go
i.Tf(ctx, "order", map[string]interface{}{
    "OrderID": "12345",
    "Name":    "John",
})
// Template: "Order #{{.OrderID}} for {{.Name}}"
// Output: "Order #12345 for John"
```

### Pluralization (Tn)

```go
i.Tn(ctx, "items", 1)  // "1 item"
i.Tn(ctx, "items", 5)  // "5 items"
```

## Pluralization Rules

The package includes CLDR-compliant pluralization rules for many languages:

- **English/German**: one (n=1), other
- **French**: one (n=0,1), other
- **Russian**: one, few, many, other (complex rules)
- **Arabic**: zero, one, two, few, many, other
- **Chinese/Japanese/Korean**: other (no plural forms)

```json
// Russian example
{
  "items": {
    "one": "{{.Count}} товар",
    "few": "{{.Count}} товара",
    "many": "{{.Count}} товаров",
    "other": "{{.Count}} товаров"
  }
}
```

## Formatting

### Number Formatting

```go
localizer := i.L("de-DE")
fmt.Println(localizer.FormatNumber(1234567.89))
// Output: 1.234.567,89

// With options
fmt.Println(localizer.FormatNumber(1234.5,
    i18n.WithMinDecimals(2),
    i18n.WithMaxDecimals(2),
))
// Output: 1.234,50
```

### Currency Formatting

```go
en := i.L("en-US")
de := i.L("de-DE")

fmt.Println(en.FormatCurrency(1234.56, "USD"))  // $1,234.56
fmt.Println(de.FormatCurrency(1234.56, "EUR"))  // 1.234,56 €

// Currency code display
fmt.Println(en.FormatCurrency(1234.56, "USD",
    i18n.WithCurrencyDisplay(i18n.CurrencyCode),
))
// Output: USD 1,234.56
```

### Date/Time Formatting

```go
t := time.Now()
en := i.L("en-US")

fmt.Println(en.FormatDate(t, i18n.DateStyleShort))   // 1/15/24
fmt.Println(en.FormatDate(t, i18n.DateStyleMedium))  // Jan 15, 2024
fmt.Println(en.FormatDate(t, i18n.DateStyleLong))    // January 15, 2024
fmt.Println(en.FormatDate(t, i18n.DateStyleFull))    // Monday, January 15, 2024

fmt.Println(en.FormatTime(t, i18n.TimeStyleShort))   // 2:30 PM
fmt.Println(en.FormatDateTime(t, i18n.DateStyleMedium, i18n.TimeStyleShort))
// Jan 15, 2024, 2:30 PM
```

### Relative Time

```go
now := time.Now()
en := i.L("en-US")

fmt.Println(en.FormatRelativeTime(now.Add(-5 * time.Second)))   // just now
fmt.Println(en.FormatRelativeTime(now.Add(-2 * time.Minute)))   // 2 minutes ago
fmt.Println(en.FormatRelativeTime(now.Add(-3 * time.Hour)))     // 3 hours ago
fmt.Println(en.FormatRelativeTime(now.Add(-24 * time.Hour)))    // yesterday
fmt.Println(en.FormatRelativeTime(now.Add(2 * time.Hour)))      // in 2 hours
```

### List Formatting

```go
items := []string{"apples", "oranges", "bananas"}
en := i.L("en-US")

fmt.Println(en.FormatList(items, i18n.ListStyleAnd))     // apples, oranges, and bananas
fmt.Println(en.FormatList(items, i18n.ListStyleOr))      // apples, oranges, or bananas
fmt.Println(en.FormatList(items, i18n.ListStyleNarrow))  // apples, oranges, bananas
```

## HTTP Middleware

```go
import (
    "github.com/rompi/core-backend/pkg/i18n"
    "github.com/rompi/core-backend/pkg/i18n/middleware"
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
        middleware.WithDefaultLocale("en"),
        middleware.WithSetCookie(true),
    )

    mux := http.NewServeMux()
    http.ListenAndServe(":8080", mw(mux))
}

// In handlers, use context
func handler(w http.ResponseWriter, r *http.Request) {
    msg := i.T(r.Context(), "welcome.message")
    fmt.Fprintf(w, msg)
}
```

## Embedded Translations

```go
import (
    "embed"
    "github.com/rompi/core-backend/pkg/i18n"
    "github.com/rompi/core-backend/pkg/i18n/catalog"
)

//go:embed locales/*.json
var localesFS embed.FS

func main() {
    cat, _ := catalog.NewEmbedCatalog(localesFS, "locales")

    i, _ := i18n.New(i18n.Config{
        DefaultLocale: "en",
    }, i18n.WithCatalog(cat))

    // Translations are embedded in binary
}
```

## RTL Support

```go
en := i.L("en")
ar := i.L("ar")
he := i.L("he")

fmt.Println(en.Direction())  // ltr
fmt.Println(ar.Direction())  // rtl
fmt.Println(he.Direction())  // rtl

// Use in templates
// <html dir="{{.Localizer.Direction}}">
```

## Custom Options

### Custom Logger

```go
i, err := i18n.New(cfg, i18n.WithLogger(myLogger))
```

### Custom Catalog

```go
i, err := i18n.New(cfg, i18n.WithCatalog(myCatalog))
```

### Missing Translation Handler

```go
i, err := i18n.New(cfg,
    i18n.WithMissingHandler(func(locale, key string) {
        log.Printf("Missing translation: %s/%s", locale, key)
        metrics.IncrCounter("i18n.missing", 1)
    }),
)
```

## Error Handling

```go
var (
    ErrLocaleNotFound = errors.New("i18n: locale not found")
    ErrKeyNotFound    = errors.New("i18n: key not found")
    ErrInvalidFormat  = errors.New("i18n: invalid format")
    ErrCatalogLoad    = errors.New("i18n: catalog load failed")
)
```

## Package Structure

```
pkg/i18n/
├── i18n.go               # Core I18n interface and implementation
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
│   └── embed.go          # Embedded FS catalog
├── middleware/
│   └── http.go           # HTTP middleware
└── examples/
    ├── basic/
    ├── http-server/
    ├── pluralization/
    └── formatting/
```

## Dependencies

- **Required:** None (core uses stdlib only)
- **Optional:**
  - `gopkg.in/yaml.v3` for YAML catalog

## License

See repository LICENSE file.
