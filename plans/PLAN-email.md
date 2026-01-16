# Package Plan: pkg/email

## Overview

A flexible email service supporting multiple providers (SMTP, SendGrid, AWS SES, Mailgun). Provides template rendering, attachments, and async sending with retry support.

## Goals

1. **Unified Interface** - Single API for all email providers
2. **Multiple Backends** - SMTP, SendGrid, AWS SES, Mailgun
3. **Template Support** - HTML/text templates with variable substitution
4. **Attachments** - Support file and inline attachments
5. **Async Sending** - Non-blocking send with callback support
6. **Retry Mechanism** - Automatic retry on transient failures
7. **Zero Core Dependencies** - Provider SDKs are optional

## Architecture

```
pkg/email/
├── email.go             # Core interfaces (Sender, Email)
├── config.go            # Configuration with env support
├── options.go           # Functional options
├── message.go           # Email message type
├── template.go          # Template engine
├── errors.go            # Custom error types
├── smtp/
│   ├── smtp.go          # SMTP implementation
│   ├── config.go
│   └── smtp_test.go
├── sendgrid/
│   ├── sendgrid.go      # SendGrid implementation
│   ├── config.go
│   └── sendgrid_test.go
├── ses/
│   ├── ses.go           # AWS SES implementation
│   ├── config.go
│   └── ses_test.go
├── mailgun/
│   ├── mailgun.go       # Mailgun implementation
│   ├── config.go
│   └── mailgun_test.go
├── mock/
│   └── mock.go          # Mock sender for testing
├── examples/
│   ├── basic/
│   ├── templates/
│   ├── attachments/
│   └── async/
└── README.md
```

## Core Interfaces

```go
package email

import (
    "context"
    "io"
)

// Sender sends emails
type Sender interface {
    // Send sends an email synchronously
    Send(ctx context.Context, email *Email) error

    // SendAsync sends an email asynchronously
    SendAsync(ctx context.Context, email *Email, callback func(error))

    // Close releases resources
    Close() error
}

// Email represents an email message
type Email struct {
    // From address
    From Address

    // To recipients
    To []Address

    // CC recipients
    CC []Address

    // BCC recipients
    BCC []Address

    // Reply-To address
    ReplyTo *Address

    // Subject line
    Subject string

    // Text body (plain text)
    Text string

    // HTML body
    HTML string

    // Attachments
    Attachments []Attachment

    // Headers for custom headers
    Headers map[string]string

    // Tags for tracking/categorization
    Tags []string

    // Metadata for provider-specific data
    Metadata map[string]string
}

// Address represents an email address
type Address struct {
    Name  string
    Email string
}

// Attachment represents an email attachment
type Attachment struct {
    // Filename
    Filename string

    // Content type (MIME)
    ContentType string

    // Content (reader or bytes)
    Content io.Reader

    // Inline for inline attachments (images in HTML)
    Inline bool

    // ContentID for inline references
    ContentID string
}

// Template renders email templates
type Template interface {
    // Render renders a template with data
    Render(name string, data interface{}) (*Email, error)

    // RenderHTML renders HTML template
    RenderHTML(name string, data interface{}) (string, error)

    // RenderText renders text template
    RenderText(name string, data interface{}) (string, error)
}
```

## Configuration

```go
// Config holds email configuration
type Config struct {
    // Backend type: "smtp", "sendgrid", "ses", "mailgun", "mock"
    Backend string `env:"EMAIL_BACKEND" default:"smtp"`

    // Default from address
    DefaultFrom Address

    // Retry configuration
    Retry RetryConfig

    // Rate limiting
    RateLimit RateLimitConfig
}

type RetryConfig struct {
    // Maximum retry attempts
    MaxAttempts int `env:"EMAIL_RETRY_MAX_ATTEMPTS" default:"3"`

    // Initial delay between retries
    InitialDelay time.Duration `env:"EMAIL_RETRY_INITIAL_DELAY" default:"1s"`

    // Maximum delay
    MaxDelay time.Duration `env:"EMAIL_RETRY_MAX_DELAY" default:"30s"`
}

type RateLimitConfig struct {
    // Enable rate limiting
    Enabled bool `env:"EMAIL_RATE_LIMIT_ENABLED" default:"false"`

    // Requests per second
    RequestsPerSecond float64 `env:"EMAIL_RATE_LIMIT_RPS" default:"10"`

    // Burst size
    Burst int `env:"EMAIL_RATE_LIMIT_BURST" default:"20"`
}
```

## Backend Configurations

### SMTP

```go
type SMTPConfig struct {
    // Host
    Host string `env:"EMAIL_SMTP_HOST" default:"localhost"`

    // Port
    Port int `env:"EMAIL_SMTP_PORT" default:"587"`

    // Username
    Username string `env:"EMAIL_SMTP_USERNAME" default:""`

    // Password
    Password string `env:"EMAIL_SMTP_PASSWORD" default:""`

    // Auth type: "plain", "login", "cram-md5"
    AuthType string `env:"EMAIL_SMTP_AUTH_TYPE" default:"plain"`

    // TLS mode: "none", "starttls", "ssl"
    TLSMode string `env:"EMAIL_SMTP_TLS_MODE" default:"starttls"`

    // Skip TLS verification (not recommended for production)
    InsecureSkipVerify bool `env:"EMAIL_SMTP_INSECURE" default:"false"`

    // Connection timeout
    ConnTimeout time.Duration `env:"EMAIL_SMTP_CONN_TIMEOUT" default:"10s"`

    // Send timeout
    SendTimeout time.Duration `env:"EMAIL_SMTP_SEND_TIMEOUT" default:"30s"`

    // Connection pool size
    PoolSize int `env:"EMAIL_SMTP_POOL_SIZE" default:"5"`
}
```

### SendGrid

```go
type SendGridConfig struct {
    // API key
    APIKey string `env:"EMAIL_SENDGRID_API_KEY" required:"true"`

    // Sandbox mode for testing
    SandboxMode bool `env:"EMAIL_SENDGRID_SANDBOX" default:"false"`

    // Enable click tracking
    ClickTracking bool `env:"EMAIL_SENDGRID_CLICK_TRACKING" default:"false"`

    // Enable open tracking
    OpenTracking bool `env:"EMAIL_SENDGRID_OPEN_TRACKING" default:"false"`

    // IP pool name
    IPPoolName string `env:"EMAIL_SENDGRID_IP_POOL" default:""`
}
```

### AWS SES

```go
type SESConfig struct {
    // Region
    Region string `env:"EMAIL_SES_REGION" default:"us-east-1"`

    // Access key (if not using IAM/environment)
    AccessKeyID     string `env:"EMAIL_SES_ACCESS_KEY_ID" default:""`
    SecretAccessKey string `env:"EMAIL_SES_SECRET_ACCESS_KEY" default:""`

    // Configuration set name
    ConfigurationSet string `env:"EMAIL_SES_CONFIGURATION_SET" default:""`

    // Endpoint (for localstack testing)
    Endpoint string `env:"EMAIL_SES_ENDPOINT" default:""`
}
```

### Mailgun

```go
type MailgunConfig struct {
    // Domain
    Domain string `env:"EMAIL_MAILGUN_DOMAIN" required:"true"`

    // API key
    APIKey string `env:"EMAIL_MAILGUN_API_KEY" required:"true"`

    // API base URL (EU: "https://api.eu.mailgun.net/v3")
    APIBase string `env:"EMAIL_MAILGUN_API_BASE" default:"https://api.mailgun.net/v3"`

    // Enable tracking
    Tracking bool `env:"EMAIL_MAILGUN_TRACKING" default:"false"`
}
```

## Template Engine

```go
// TemplateEngine provides template rendering
type TemplateEngine struct {
    templates *template.Template
    funcMap   template.FuncMap
}

// NewTemplateEngine creates a template engine
func NewTemplateEngine(opts ...TemplateOption) *TemplateEngine

// TemplateOption configures the template engine
type TemplateOption func(*TemplateEngine)

// WithTemplateDir loads templates from directory
func WithTemplateDir(dir string) TemplateOption

// WithTemplateFS loads templates from embed.FS
func WithTemplateFS(fs embed.FS, pattern string) TemplateOption

// WithFuncMap adds custom template functions
func WithFuncMap(funcs template.FuncMap) TemplateOption
```

### Built-in Template Functions

```go
// Default template functions
var DefaultFuncMap = template.FuncMap{
    "formatDate":     formatDate,     // {{formatDate .Date "Jan 02, 2006"}}
    "formatCurrency": formatCurrency, // {{formatCurrency .Amount "USD"}}
    "truncate":       truncate,       // {{truncate .Text 100}}
    "htmlSafe":       htmlSafe,       // {{htmlSafe .HTML}}
    "urlEncode":      urlEncode,      // {{urlEncode .Query}}
    "upper":          strings.ToUpper,
    "lower":          strings.ToLower,
    "title":          strings.Title,
}
```

## Email Builder

```go
// Builder provides fluent email construction
type Builder struct {
    email *Email
}

// New creates a new email builder
func New() *Builder

// From sets sender
func (b *Builder) From(name, email string) *Builder

// To adds recipient
func (b *Builder) To(name, email string) *Builder

// CC adds CC recipient
func (b *Builder) CC(name, email string) *Builder

// BCC adds BCC recipient
func (b *Builder) BCC(name, email string) *Builder

// ReplyTo sets reply-to address
func (b *Builder) ReplyTo(name, email string) *Builder

// Subject sets subject line
func (b *Builder) Subject(subject string) *Builder

// Text sets plain text body
func (b *Builder) Text(text string) *Builder

// HTML sets HTML body
func (b *Builder) HTML(html string) *Builder

// Attach adds attachment
func (b *Builder) Attach(filename string, content io.Reader, contentType string) *Builder

// AttachFile adds file attachment
func (b *Builder) AttachFile(path string) *Builder

// InlineImage adds inline image
func (b *Builder) InlineImage(contentID string, content io.Reader, contentType string) *Builder

// Header adds custom header
func (b *Builder) Header(key, value string) *Builder

// Tag adds tag
func (b *Builder) Tag(tag string) *Builder

// Build creates the email
func (b *Builder) Build() *Email
```

## Error Handling

```go
var (
    // ErrInvalidEmail is returned for invalid email addresses
    ErrInvalidEmail = errors.New("email: invalid email address")

    // ErrNoRecipients is returned when no recipients specified
    ErrNoRecipients = errors.New("email: no recipients specified")

    // ErrSendFailed is returned when sending fails
    ErrSendFailed = errors.New("email: send failed")

    // ErrRateLimited is returned when rate limited
    ErrRateLimited = errors.New("email: rate limited")

    // ErrTemplateNotFound is returned for missing templates
    ErrTemplateNotFound = errors.New("email: template not found")

    // ErrAuthFailed is returned on authentication failure
    ErrAuthFailed = errors.New("email: authentication failed")
)

// SendError provides detailed send failure info
type SendError struct {
    Err        error
    StatusCode int
    Message    string
    Retryable  bool
}
```

## Usage Examples

### Basic Email

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/email"
    "github.com/user/core-backend/pkg/email/smtp"
)

func main() {
    // Create SMTP sender
    sender, err := smtp.New(smtp.Config{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "user@gmail.com",
        Password: "app-password",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sender.Close()

    ctx := context.Background()

    // Send email
    err = sender.Send(ctx, &email.Email{
        From:    email.Address{Name: "My App", Email: "noreply@myapp.com"},
        To:      []email.Address{{Name: "John", Email: "john@example.com"}},
        Subject: "Welcome to My App",
        Text:    "Thanks for signing up!",
        HTML:    "<h1>Welcome!</h1><p>Thanks for signing up!</p>",
    })
}
```

### Using Builder

```go
func main() {
    sender, _ := smtp.New(cfg)

    mail := email.New().
        From("My App", "noreply@myapp.com").
        To("John", "john@example.com").
        Subject("Welcome!").
        HTML("<h1>Welcome</h1>").
        Text("Welcome").
        Build()

    sender.Send(ctx, mail)
}
```

### With Templates

```go
//go:embed templates/*
var templateFS embed.FS

func main() {
    // Create template engine
    tmpl := email.NewTemplateEngine(
        email.WithTemplateFS(templateFS, "templates/*.html"),
        email.WithFuncMap(template.FuncMap{
            "appName": func() string { return "MyApp" },
        }),
    )

    // Render template
    mail, err := tmpl.Render("welcome", map[string]interface{}{
        "UserName": "John",
        "VerifyURL": "https://myapp.com/verify?token=abc",
    })

    sender.Send(ctx, mail)
}

// templates/welcome.html
/*
{{define "subject"}}Welcome to {{appName}}, {{.UserName}}!{{end}}

{{define "html"}}
<!DOCTYPE html>
<html>
<body>
    <h1>Welcome, {{.UserName}}!</h1>
    <p>Please verify your email:</p>
    <a href="{{.VerifyURL}}">Verify Email</a>
</body>
</html>
{{end}}

{{define "text"}}
Welcome, {{.UserName}}!

Please verify your email: {{.VerifyURL}}
{{end}}
*/
```

### With Attachments

```go
func main() {
    sender, _ := smtp.New(cfg)

    // Open file
    file, _ := os.Open("report.pdf")
    defer file.Close()

    mail := email.New().
        From("Reports", "reports@myapp.com").
        To("Manager", "manager@company.com").
        Subject("Monthly Report").
        Text("Please find the monthly report attached.").
        Attach("report.pdf", file, "application/pdf").
        Build()

    sender.Send(ctx, mail)
}
```

### Inline Images

```go
func main() {
    sender, _ := smtp.New(cfg)

    logo, _ := os.Open("logo.png")
    defer logo.Close()

    mail := email.New().
        From("My App", "noreply@myapp.com").
        To("John", "john@example.com").
        Subject("Newsletter").
        HTML(`<html><body><img src="cid:logo"><p>Content here</p></body></html>`).
        InlineImage("logo", logo, "image/png").
        Build()

    sender.Send(ctx, mail)
}
```

### SendGrid

```go
import (
    "github.com/user/core-backend/pkg/email/sendgrid"
)

func main() {
    sender, err := sendgrid.New(sendgrid.Config{
        APIKey: os.Getenv("SENDGRID_API_KEY"),
    })

    // Same interface
    sender.Send(ctx, mail)
}
```

### Async Sending

```go
func main() {
    sender, _ := smtp.New(cfg)

    // Send asynchronously
    sender.SendAsync(ctx, mail, func(err error) {
        if err != nil {
            log.Printf("Email failed: %v", err)
        } else {
            log.Println("Email sent successfully")
        }
    })

    // Continue execution
}
```

### Testing with Mock

```go
import (
    "github.com/user/core-backend/pkg/email/mock"
)

func TestSendWelcomeEmail(t *testing.T) {
    // Create mock sender
    sender := mock.New()

    // Send email
    err := SendWelcomeEmail(sender, user)
    require.NoError(t, err)

    // Verify
    require.Len(t, sender.Sent, 1)
    assert.Equal(t, "Welcome!", sender.Sent[0].Subject)
    assert.Contains(t, sender.Sent[0].HTML, user.Name)
}
```

## Common Templates

```go
// Package email provides common email templates
package email

// WelcomeEmail creates a welcome email
func WelcomeEmail(userName, verifyURL string) *Email

// PasswordResetEmail creates a password reset email
func PasswordResetEmail(userName, resetURL string) *Email

// InvoiceEmail creates an invoice email
func InvoiceEmail(invoiceNumber string, amount float64, dueDate time.Time) *Email

// NotificationEmail creates a generic notification
func NotificationEmail(title, message string) *Email
```

## Health Check

```go
// HealthCheck verifies email service connectivity
func (s *SMTP) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        conn, err := s.dial(ctx)
        if err != nil {
            return err
        }
        return conn.Close()
    }
}
```

## Dependencies

- **Required:** None (SMTP uses stdlib)
- **Optional:**
  - `github.com/sendgrid/sendgrid-go` for SendGrid
  - `github.com/aws/aws-sdk-go-v2/service/ses` for AWS SES
  - `github.com/mailgun/mailgun-go` for Mailgun

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with MailHog/Mailpit for SMTP
- Template rendering tests
- Attachment encoding tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & SMTP Implementation
1. Define Sender, Email interfaces
2. Implement SMTP sender with connection pooling
3. Add email builder
4. Write comprehensive tests

### Phase 2: Template Engine
1. Implement template engine
2. Built-in template functions
3. Common email templates
4. Template tests

### Phase 3: SendGrid Implementation
1. Implement SendGrid sender
2. Tracking options
3. Integration tests

### Phase 4: AWS SES Implementation
1. Implement SES sender
2. Configuration set support
3. Integration tests

### Phase 5: Mailgun Implementation
1. Implement Mailgun sender
2. Tracking options
3. Integration tests

### Phase 6: Advanced Features
1. Async sending with callbacks
2. Rate limiting
3. Retry mechanism
4. Mock sender for testing

### Phase 7: Documentation & Examples
1. README with full documentation
2. Examples for each provider
3. Template examples
4. Testing examples
