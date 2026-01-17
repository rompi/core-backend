# Package Plan: pkg/validator

## Overview

A comprehensive input validation package with struct tag-based validation, custom validators, and localized error messages. Designed for validating API requests, form data, and configuration with clear, actionable error messages.

## Goals

1. **Struct Tag Validation** - Declarative validation via struct tags
2. **Custom Validators** - Extensible with custom validation rules
3. **Localized Errors** - Multi-language error messages
4. **Nested Validation** - Deep validation of nested structs/slices
5. **Conditional Validation** - Rules based on other field values
6. **HTTP Integration** - Middleware for request validation
7. **Zero External Dependencies** - Pure Go implementation

## Architecture

```
pkg/validator/
├── validator.go          # Core Validator interface
├── config.go             # Configuration
├── options.go            # Functional options
├── rules.go              # Built-in validation rules
├── errors.go             # Validation errors
├── tags.go               # Struct tag parser
├── types.go              # Type-specific validators
├── custom.go             # Custom validator registration
├── i18n/
│   ├── i18n.go           # Internationalization
│   ├── en.go             # English messages
│   ├── es.go             # Spanish messages
│   └── ...
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── examples/
│   ├── basic/
│   ├── custom-rules/
│   ├── http-middleware/
│   └── i18n/
└── README.md
```

## Core Interfaces

```go
package validator

import (
    "context"
)

// Validator validates structs and values
type Validator interface {
    // Validate validates a struct
    Validate(ctx context.Context, v interface{}) error

    // ValidatePartial validates specific fields only
    ValidatePartial(ctx context.Context, v interface{}, fields ...string) error

    // Var validates a single variable
    Var(ctx context.Context, field interface{}, tag string) error

    // RegisterValidation registers a custom validation
    RegisterValidation(tag string, fn ValidationFunc, opts ...RuleOption) error

    // RegisterAlias registers an alias for multiple validations
    RegisterAlias(alias, tags string) error

    // RegisterStructValidation registers struct-level validation
    RegisterStructValidation(fn StructValidationFunc, types ...interface{})

    // RegisterTranslation registers custom error messages
    RegisterTranslation(tag, message string, translator TranslateFunc)
}

// ValidationFunc is a custom validation function
type ValidationFunc func(ctx context.Context, fl FieldLevel) bool

// StructValidationFunc is a struct-level validation function
type StructValidationFunc func(ctx context.Context, sl StructLevel)

// FieldLevel provides field information during validation
type FieldLevel interface {
    // Field returns the current field value
    Field() reflect.Value

    // FieldName returns the field name
    FieldName() string

    // StructFieldName returns the struct field name
    StructFieldName() string

    // Param returns the validation parameter
    Param() string

    // Parent returns the parent struct
    Parent() reflect.Value

    // Top returns the top-level struct
    Top() reflect.Value

    // GetTag returns a custom tag value
    GetTag(tag string) string
}

// StructLevel provides struct-level validation info
type StructLevel interface {
    // Current returns the current struct
    Current() reflect.Value

    // ReportError reports a validation error
    ReportError(field interface{}, fieldName, tag, param, message string)
}

// TranslateFunc translates error messages
type TranslateFunc func(fe FieldError, param string) string
```

## Validation Errors

```go
// ValidationError contains all validation errors
type ValidationError struct {
    Errors []FieldError
}

func (e ValidationError) Error() string

// HasErrors returns true if there are errors
func (e ValidationError) HasErrors() bool

// FieldErrors returns errors for a specific field
func (e ValidationError) FieldErrors(field string) []FieldError

// ToMap converts errors to map format
func (e ValidationError) ToMap() map[string][]string

// ToJSON returns JSON representation
func (e ValidationError) ToJSON() ([]byte, error)

// FieldError represents a single field validation error
type FieldError interface {
    // Field returns the field name
    Field() string

    // StructField returns the struct field name
    StructField() string

    // Tag returns the validation tag that failed
    Tag() string

    // Param returns the validation parameter
    Param() string

    // Value returns the field value
    Value() interface{}

    // Kind returns the reflect.Kind
    Kind() reflect.Kind

    // Type returns the reflect.Type
    Type() reflect.Type

    // Namespace returns the full field path
    Namespace() string

    // Error returns the error message
    Error() string

    // Translate returns localized error message
    Translate(translator Translator) string
}
```

## Built-in Validation Rules

```go
// String validations
required        // Field must be present and non-empty
min=n           // Minimum length
max=n           // Maximum length
len=n           // Exact length
email           // Valid email format
url             // Valid URL format
uri             // Valid URI format
alpha           // Alphabetic characters only
alphanum        // Alphanumeric characters only
alphanumspace   // Alphanumeric + spaces
ascii           // ASCII characters only
lowercase       // Lowercase only
uppercase       // Uppercase only
contains=s      // Contains substring
startswith=s    // Starts with prefix
endswith=s      // Ends with suffix
uuid            // Valid UUID format
uuid3           // UUID version 3
uuid4           // UUID version 4
uuid5           // UUID version 5
json            // Valid JSON string
jwt             // Valid JWT format
base64          // Valid base64 string
html            // Contains HTML (for sanitization warnings)
nohtml          // No HTML allowed

// Numeric validations
gt=n            // Greater than
gte=n           // Greater than or equal
lt=n            // Less than
lte=n           // Less than or equal
eq=n            // Equal to
ne=n            // Not equal to
positive        // Greater than 0
negative        // Less than 0
between=a,b     // Between a and b (inclusive)
oneof=a b c     // One of the values

// Time validations
datetime=layout // Valid datetime with layout
date            // Valid date (2006-01-02)
time            // Valid time (15:04:05)
timezone        // Valid timezone
before=time     // Before specific time
after=time      // After specific time
future          // In the future
past            // In the past

// Network validations
ip              // Valid IP address
ipv4            // Valid IPv4 address
ipv6            // Valid IPv6 address
cidr            // Valid CIDR notation
mac             // Valid MAC address
hostname        // Valid hostname
fqdn            // Fully qualified domain name

// File validations
file            // Existing file path
dir             // Existing directory path
filepath        // Valid file path syntax

// Format validations
creditcard      // Valid credit card number (Luhn)
ssn             // Social Security Number
isbn            // ISBN-10 or ISBN-13
isbn10          // ISBN-10
isbn13          // ISBN-13
phone           // Phone number (E.164)
postcode=CC     // Postal code for country

// Comparison validations
eqfield=Field   // Equal to another field
nefield=Field   // Not equal to another field
gtfield=Field   // Greater than another field
gtefield=Field  // Greater than or equal to another field
ltfield=Field   // Less than another field
ltefield=Field  // Less than or equal to another field

// Slice/Map validations
unique          // All elements unique
min=n           // Minimum elements
max=n           // Maximum elements
dive            // Validate each element

// Conditional validations
required_if=Field Value     // Required if field equals value
required_unless=Field Value // Required unless field equals value
required_with=Field         // Required if field is present
required_without=Field      // Required if field is absent
omitempty                   // Skip if empty
```

## Configuration

```go
// Config holds validator configuration
type Config struct {
    // Tag name for validation rules
    TagName string `default:"validate"`

    // Default language for errors
    Language string `default:"en"`

    // Fail fast on first error
    FailFast bool `default:"false"`

    // Required fields fail on zero values
    RequiredZero bool `default:"true"`
}
```

## Usage Examples

### Basic Validation

```go
package main

import (
    "context"
    "fmt"
    "github.com/user/core-backend/pkg/validator"
)

type CreateUserRequest struct {
    Name     string `validate:"required,min=2,max=100"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8,max=72"`
    Age      int    `validate:"gte=18,lte=120"`
    Website  string `validate:"omitempty,url"`
}

func main() {
    v := validator.New()
    ctx := context.Background()

    req := CreateUserRequest{
        Name:     "J", // Too short
        Email:    "invalid-email",
        Password: "short",
        Age:      15, // Too young
    }

    err := v.Validate(ctx, &req)
    if err != nil {
        if verr, ok := err.(*validator.ValidationError); ok {
            for _, fe := range verr.Errors {
                fmt.Printf("%s: %s\n", fe.Field(), fe.Error())
            }
        }
    }
}

// Output:
// Name: must be at least 2 characters
// Email: must be a valid email address
// Password: must be at least 8 characters
// Age: must be 18 or greater
```

### Nested Structs

```go
type Address struct {
    Street  string `validate:"required,max=200"`
    City    string `validate:"required,max=100"`
    Country string `validate:"required,iso3166_1_alpha2"`
    ZipCode string `validate:"required,postcode=US"`
}

type Order struct {
    ID              string    `validate:"required,uuid4"`
    CustomerEmail   string    `validate:"required,email"`
    ShippingAddress Address   `validate:"required"`
    BillingAddress  *Address  `validate:"omitempty"`
    Items           []OrderItem `validate:"required,min=1,dive"`
}

type OrderItem struct {
    ProductID string  `validate:"required,uuid4"`
    Quantity  int     `validate:"required,gte=1,lte=100"`
    Price     float64 `validate:"required,gt=0"`
}

func main() {
    v := validator.New()

    order := Order{
        ID:            "not-a-uuid",
        CustomerEmail: "customer@example.com",
        ShippingAddress: Address{
            Street: "123 Main St",
            // Missing City and Country
        },
        Items: []OrderItem{
            {ProductID: "prod-1", Quantity: 0, Price: -10}, // Invalid
        },
    }

    err := v.Validate(ctx, &order)
    // Errors for: ID, ShippingAddress.City, ShippingAddress.Country,
    // Items[0].ProductID, Items[0].Quantity, Items[0].Price
}
```

### Custom Validators

```go
func main() {
    v := validator.New()

    // Register custom validation
    v.RegisterValidation("password_strength", func(ctx context.Context, fl validator.FieldLevel) bool {
        password := fl.Field().String()

        hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
        hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
        hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
        hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)

        return hasUpper && hasLower && hasNumber && hasSpecial
    })

    // Register custom error message
    v.RegisterTranslation("password_strength",
        "must contain uppercase, lowercase, number, and special character",
        nil,
    )

    type User struct {
        Password string `validate:"required,min=8,password_strength"`
    }

    v.Validate(ctx, &User{Password: "weakpassword"})
}
```

### Struct-Level Validation

```go
type DateRange struct {
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required"`
}

func main() {
    v := validator.New()

    // Register struct-level validation
    v.RegisterStructValidation(func(ctx context.Context, sl validator.StructLevel) {
        dr := sl.Current().Interface().(DateRange)

        if !dr.EndDate.After(dr.StartDate) {
            sl.ReportError(dr.EndDate, "EndDate", "dateafter", "",
                "end date must be after start date")
        }
    }, DateRange{})

    dr := DateRange{
        StartDate: time.Now(),
        EndDate:   time.Now().Add(-24 * time.Hour), // Before start
    }

    v.Validate(ctx, &dr)
}
```

### Conditional Validation

```go
type Payment struct {
    Method       string `validate:"required,oneof=card bank crypto"`
    CardNumber   string `validate:"required_if=Method card,omitempty,creditcard"`
    CardExpiry   string `validate:"required_if=Method card,omitempty,datetime=01/06"`
    CardCVV      string `validate:"required_if=Method card,omitempty,len=3"`
    BankAccount  string `validate:"required_if=Method bank,omitempty,alphanum"`
    BankRouting  string `validate:"required_if=Method bank,omitempty,len=9"`
    CryptoWallet string `validate:"required_if=Method crypto,omitempty,alphanum"`
}

func main() {
    v := validator.New()

    payment := Payment{
        Method:     "card",
        CardNumber: "", // Required when Method=card
    }

    v.Validate(ctx, &payment)
    // Error: CardNumber is required when Method is card
}
```

### Slice/Map Validation

```go
type Config struct {
    // Each tag must be unique and alphanumeric
    Tags []string `validate:"required,min=1,max=10,unique,dive,required,alphanum,max=20"`

    // Each setting key must be valid, value must be non-empty
    Settings map[string]string `validate:"required,dive,keys,alphanum,endkeys,required,max=100"`

    // Nested slice validation
    Users []User `validate:"required,min=1,dive"`
}

func main() {
    v := validator.New()

    cfg := Config{
        Tags: []string{"tag1", "tag1", ""}, // Duplicate and empty
        Settings: map[string]string{
            "valid-key": "", // Empty value
        },
    }

    v.Validate(ctx, &cfg)
}
```

### Internationalization

```go
import (
    "github.com/user/core-backend/pkg/validator"
    "github.com/user/core-backend/pkg/validator/i18n"
)

func main() {
    v := validator.New()

    // Load Spanish translations
    i18n.RegisterSpanish(v)

    ctx := context.WithValue(context.Background(), validator.LanguageKey, "es")

    type User struct {
        Email string `validate:"required,email"`
    }

    err := v.Validate(ctx, &User{Email: "invalid"})
    if verr, ok := err.(*validator.ValidationError); ok {
        for _, fe := range verr.Errors {
            // Error message in Spanish
            fmt.Println(fe.Error()) // "debe ser una dirección de correo válida"
        }
    }
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/validator"
    "github.com/user/core-backend/pkg/validator/middleware"
)

func main() {
    v := validator.New()

    mux := http.NewServeMux()

    // Apply validation middleware
    handler := middleware.ValidateJSON(v, CreateUserRequest{})(
        http.HandlerFunc(createUserHandler),
    )

    mux.Handle("POST /users", handler)
}

// Or manual validation in handler
func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := v.Validate(r.Context(), &req); err != nil {
        if verr, ok := err.(*validator.ValidationError); ok {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "errors": verr.ToMap(),
            })
            return
        }
    }

    // Process valid request
}
```

### gRPC Interceptor

```go
import (
    "github.com/user/core-backend/pkg/validator/middleware"
)

func main() {
    v := validator.New()

    server := grpc.NewServer(
        grpc.UnaryInterceptor(middleware.UnaryValidateInterceptor(v)),
    )
}
```

### Partial Validation

```go
type UpdateUserRequest struct {
    Name     string `validate:"omitempty,min=2,max=100"`
    Email    string `validate:"omitempty,email"`
    Password string `validate:"omitempty,min=8,max=72"`
}

func main() {
    v := validator.New()

    req := UpdateUserRequest{
        Name: "J", // Too short, but we only validate Email
    }

    // Only validate Email field
    err := v.ValidatePartial(ctx, &req, "Email")
    // No error because Email is valid (empty with omitempty)
}
```

### Single Variable Validation

```go
func main() {
    v := validator.New()

    email := "test@example"

    err := v.Var(ctx, email, "required,email")
    if err != nil {
        fmt.Println("Invalid email")
    }
}
```

## Error Response Format

```go
// ToMap returns field -> errors map
verr.ToMap()
// {
//   "Email": ["must be a valid email address"],
//   "Password": ["must be at least 8 characters", "must contain a number"]
// }

// ToJSON returns JSON format
verr.ToJSON()
// {
//   "errors": [
//     {"field": "Email", "tag": "email", "message": "must be a valid email address"},
//     {"field": "Password", "tag": "min", "param": "8", "message": "must be at least 8 characters"}
//   ]
// }
```

## Dependencies

- **Required:** None (uses reflect from stdlib)
- **Optional:** None

## Test Coverage Requirements

- Unit tests for all built-in rules
- Edge case tests (empty, nil, zero values)
- Nested struct tests
- Custom validator tests
- i18n tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Validator & Basic Rules
1. Define Validator interface
2. Struct tag parser
3. Basic string validations (required, min, max, email, etc.)
4. Numeric validations (gt, gte, lt, lte)
5. Validation error types

### Phase 2: Advanced Rules
1. Time validations
2. Network validations
3. Format validations (creditcard, uuid, etc.)
4. Comparison validations (eqfield, gtfield, etc.)

### Phase 3: Nested & Collection Validation
1. Nested struct validation
2. Slice validation with dive
3. Map validation
4. Pointer handling

### Phase 4: Custom Validators
1. Custom validation registration
2. Struct-level validation
3. Alias registration

### Phase 5: Internationalization
1. Translation interface
2. English messages
3. Additional languages
4. Context-based language selection

### Phase 6: Middleware & Integration
1. HTTP middleware
2. gRPC interceptor
3. Error response formatting

### Phase 7: Documentation & Examples
1. README with full documentation
2. Basic example
3. Custom rules example
4. i18n example
