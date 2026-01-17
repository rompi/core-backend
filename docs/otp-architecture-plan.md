# OTP Package Architecture Plan

> A plug-and-play, provider-agnostic One-Time Password (OTP) package for the core-backend library.

## Overview

This document outlines the architecture for an OTP package that follows the established patterns in this codebase:
- **Interface-based dependency injection** for storage and notification providers
- **Functional options pattern** for extensibility
- **Environment-driven configuration** with sensible defaults
- **Zero coupling** to specific implementations

The OTP package will support multiple OTP types (TOTP, HOTP, numeric codes for SMS/Email) while remaining agnostic of how codes are stored or delivered.

---

## Package Structure

```
pkg/otp/
├── service.go          # Core OTP service
├── config.go           # Configuration with env loading
├── models.go           # OTP data models
├── repository.go       # Storage interface
├── notifier.go         # Notification interface
├── generator.go        # OTP generation strategies
├── options.go          # Functional options
├── errors.go           # Package-specific errors
├── logger.go           # Logger interface + NoopLogger
├── middleware.go       # HTTP/gRPC middleware helpers
├── testutil/
│   └── mocks.go        # Mock implementations for testing
└── examples/
    ├── basic/          # Simple OTP flow
    ├── totp/           # TOTP (authenticator app) example
    ├── sms/            # SMS OTP example
    └── email/          # Email OTP example
```

---

## Core Interfaces

### 1. OTP Repository Interface (Storage Agnostic)

```go
// OTPRepository defines the storage contract for OTP tokens.
// Implementations can use any storage backend: PostgreSQL, Redis, MongoDB, in-memory, etc.
type OTPRepository interface {
    // Create stores a new OTP token
    Create(ctx context.Context, otp *OTP) error

    // GetByID retrieves an OTP by its unique identifier
    GetByID(ctx context.Context, id string) (*OTP, error)

    // GetByUserAndPurpose retrieves the latest OTP for a user and purpose
    GetByUserAndPurpose(ctx context.Context, userID string, purpose Purpose) (*OTP, error)

    // Update updates an existing OTP (e.g., increment attempts, mark as used)
    Update(ctx context.Context, otp *OTP) error

    // Delete removes an OTP by ID
    Delete(ctx context.Context, id string) error

    // DeleteByUserAndPurpose removes all OTPs for a user and purpose
    DeleteByUserAndPurpose(ctx context.Context, userID string, purpose Purpose) error

    // DeleteExpired removes all expired OTPs (for cleanup jobs)
    DeleteExpired(ctx context.Context) (int64, error)
}
```

### 2. Notifier Interface (Delivery Agnostic)

```go
// Notifier defines the contract for OTP delivery.
// Implementations can use any delivery mechanism: SMS, Email, Push, etc.
type Notifier interface {
    // Send delivers the OTP to the user via the configured channel
    Send(ctx context.Context, request *NotificationRequest) error
}

// NotificationRequest contains all information needed to deliver an OTP
type NotificationRequest struct {
    // Recipient identifier (email, phone number, device token, etc.)
    Recipient string

    // The OTP code to deliver
    Code string

    // Purpose of the OTP (for message customization)
    Purpose Purpose

    // Channel hint (email, sms, push) - notifier may ignore if single-channel
    Channel Channel

    // Additional metadata for template rendering
    Metadata map[string]string

    // Locale for message localization
    Locale string
}

// Channel represents the delivery channel
type Channel string

const (
    ChannelEmail Channel = "email"
    ChannelSMS   Channel = "sms"
    ChannelPush  Channel = "push"
    ChannelAuto  Channel = "auto" // Let notifier decide
)
```

### 3. Generator Interface (Algorithm Agnostic)

```go
// Generator defines the contract for OTP code generation.
// Implementations can use different algorithms: random numeric, TOTP, HOTP, etc.
type Generator interface {
    // Generate creates a new OTP code
    Generate(ctx context.Context, params *GenerateParams) (string, error)

    // Validate checks if a code is valid (for TOTP/HOTP with time/counter windows)
    Validate(ctx context.Context, params *ValidateParams) (bool, error)
}

// GenerateParams contains parameters for OTP generation
type GenerateParams struct {
    // Secret key (for TOTP/HOTP)
    Secret string

    // Counter value (for HOTP)
    Counter uint64

    // Length of the OTP code
    Length int

    // Algorithm type
    Algorithm Algorithm
}

// ValidateParams contains parameters for OTP validation
type ValidateParams struct {
    // The code to validate
    Code string

    // Secret key (for TOTP/HOTP)
    Secret string

    // Counter value (for HOTP)
    Counter uint64

    // Algorithm type
    Algorithm Algorithm

    // Time window for TOTP (number of periods to check)
    Window int
}

// Algorithm represents the OTP algorithm type
type Algorithm string

const (
    AlgorithmNumeric Algorithm = "numeric"  // Random numeric code
    AlgorithmTOTP    Algorithm = "totp"     // Time-based OTP
    AlgorithmHOTP    Algorithm = "hotp"     // HMAC-based OTP
)
```

### 4. Audit Logger Interface (Optional)

```go
// AuditLogger defines the contract for OTP audit logging.
// This is optional - if not provided, audit logging is skipped.
type AuditLogger interface {
    // Log records an OTP-related event
    Log(ctx context.Context, event *AuditEvent) error
}

// AuditEvent represents an OTP audit log entry
type AuditEvent struct {
    Timestamp   time.Time
    UserID      string
    Action      AuditAction
    Purpose     Purpose
    Success     bool
    IPAddress   string
    UserAgent   string
    Metadata    map[string]string
}

// AuditAction represents the type of OTP action
type AuditAction string

const (
    AuditActionGenerate      AuditAction = "otp.generate"
    AuditActionVerify        AuditAction = "otp.verify"
    AuditActionVerifyFailed  AuditAction = "otp.verify_failed"
    AuditActionExpired       AuditAction = "otp.expired"
    AuditActionMaxAttempts   AuditAction = "otp.max_attempts"
    AuditActionResend        AuditAction = "otp.resend"
)
```

---

## Data Models

```go
// OTP represents a one-time password token
type OTP struct {
    // Unique identifier
    ID string

    // User this OTP belongs to
    UserID string

    // The OTP code (hashed for storage, or empty for TOTP)
    CodeHash string

    // Purpose of the OTP
    Purpose Purpose

    // Algorithm used
    Algorithm Algorithm

    // Secret key (for TOTP/HOTP, encrypted at rest)
    Secret string

    // Counter (for HOTP)
    Counter uint64

    // Number of verification attempts
    Attempts int

    // Maximum allowed attempts
    MaxAttempts int

    // Whether the OTP has been used
    Used bool

    // Creation timestamp
    CreatedAt time.Time

    // Expiration timestamp
    ExpiresAt time.Time

    // Last verification attempt timestamp
    LastAttemptAt *time.Time

    // Metadata for additional context
    Metadata map[string]string
}

// Purpose defines the reason for the OTP
type Purpose string

const (
    PurposeLogin           Purpose = "login"
    PurposeRegistration    Purpose = "registration"
    PurposePasswordReset   Purpose = "password_reset"
    PurposeEmailVerify     Purpose = "email_verify"
    PurposePhoneVerify     Purpose = "phone_verify"
    PurposeTransaction     Purpose = "transaction"
    PurposeTwoFactor       Purpose = "two_factor"
    PurposeCustom          Purpose = "custom"
)

// TOTPSetup contains information for TOTP setup (authenticator apps)
type TOTPSetup struct {
    // Secret key (base32 encoded)
    Secret string

    // QR code URL for authenticator apps
    QRCodeURL string

    // Manual entry key
    ManualEntryKey string

    // Issuer name
    Issuer string

    // Account name (usually email)
    AccountName string

    // Recovery codes (for backup)
    RecoveryCodes []string
}
```

---

## Configuration

```go
// Config holds the OTP service configuration
type Config struct {
    // Default OTP code length (default: 6)
    CodeLength int

    // Default TTL for OTP tokens (default: 5 minutes)
    TTL time.Duration

    // Maximum verification attempts before lockout (default: 3)
    MaxAttempts int

    // Cooldown period between OTP generation requests (default: 60 seconds)
    ResendCooldown time.Duration

    // Default algorithm (default: numeric)
    DefaultAlgorithm Algorithm

    // TOTP configuration
    TOTP TOTPConfig

    // Rate limiting configuration
    RateLimit RateLimitConfig

    // Whether to hash OTP codes before storage (default: true)
    HashCodes bool

    // Hash algorithm for code hashing (default: sha256)
    HashAlgorithm string
}

// TOTPConfig holds TOTP-specific configuration
type TOTPConfig struct {
    // Issuer name for authenticator apps
    Issuer string

    // Time step in seconds (default: 30)
    Period int

    // Number of periods to check for clock skew (default: 1)
    Skew int

    // Number of recovery codes to generate (default: 10)
    RecoveryCodeCount int

    // Length of recovery codes (default: 8)
    RecoveryCodeLength int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
    // Enable rate limiting (default: true)
    Enabled bool

    // Maximum OTP generation requests per user per hour (default: 10)
    MaxGeneratePerHour int

    // Maximum verification attempts per OTP (default: 3)
    MaxVerifyAttempts int

    // Lockout duration after max attempts (default: 15 minutes)
    LockoutDuration time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
    return &Config{
        CodeLength:       6,
        TTL:              5 * time.Minute,
        MaxAttempts:      3,
        ResendCooldown:   60 * time.Second,
        DefaultAlgorithm: AlgorithmNumeric,
        HashCodes:        true,
        HashAlgorithm:    "sha256",
        TOTP: TOTPConfig{
            Issuer:             "MyApp",
            Period:             30,
            Skew:               1,
            RecoveryCodeCount:  10,
            RecoveryCodeLength: 8,
        },
        RateLimit: RateLimitConfig{
            Enabled:            true,
            MaxGeneratePerHour: 10,
            MaxVerifyAttempts:  3,
            LockoutDuration:    15 * time.Minute,
        },
    }
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    cfg := DefaultConfig()

    // OTP_CODE_LENGTH
    // OTP_TTL
    // OTP_MAX_ATTEMPTS
    // OTP_RESEND_COOLDOWN
    // OTP_DEFAULT_ALGORITHM
    // OTP_HASH_CODES
    // OTP_TOTP_ISSUER
    // OTP_TOTP_PERIOD
    // OTP_TOTP_SKEW
    // OTP_RATE_LIMIT_ENABLED
    // OTP_RATE_LIMIT_MAX_GENERATE_PER_HOUR
    // OTP_RATE_LIMIT_LOCKOUT_DURATION

    return cfg, cfg.Validate()
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
    if c.CodeLength < 4 || c.CodeLength > 10 {
        return errors.New("code length must be between 4 and 10")
    }
    if c.TTL < time.Minute || c.TTL > time.Hour {
        return errors.New("TTL must be between 1 minute and 1 hour")
    }
    if c.MaxAttempts < 1 || c.MaxAttempts > 10 {
        return errors.New("max attempts must be between 1 and 10")
    }
    return nil
}
```

---

## Service API

```go
// Service provides OTP functionality
type Service struct {
    config      *Config
    repository  OTPRepository
    notifier    Notifier      // Optional
    generator   Generator
    auditLogger AuditLogger   // Optional
    logger      Logger
}

// NewService creates a new OTP service
func NewService(cfg *Config, repository OTPRepository, opts ...Option) (*Service, error) {
    if cfg == nil {
        cfg = DefaultConfig()
    }
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    if repository == nil {
        return nil, errors.New("repository is required")
    }

    s := &Service{
        config:     cfg,
        repository: repository,
        generator:  NewDefaultGenerator(), // Built-in generator
        logger:     &NoopLogger{},
    }

    for _, opt := range opts {
        if err := opt(s); err != nil {
            return nil, err
        }
    }

    return s, nil
}

// Generate creates a new OTP for a user
func (s *Service) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

// Verify validates an OTP code
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)

// Resend generates and sends a new OTP (respecting cooldown)
func (s *Service) Resend(ctx context.Context, req *ResendRequest) (*GenerateResponse, error)

// SetupTOTP initializes TOTP for a user (authenticator apps)
func (s *Service) SetupTOTP(ctx context.Context, req *SetupTOTPRequest) (*TOTPSetup, error)

// VerifyTOTP validates a TOTP code
func (s *Service) VerifyTOTP(ctx context.Context, req *VerifyTOTPRequest) (*VerifyResponse, error)

// EnableTOTP enables TOTP for a user after successful verification
func (s *Service) EnableTOTP(ctx context.Context, userID string, code string) error

// DisableTOTP disables TOTP for a user
func (s *Service) DisableTOTP(ctx context.Context, userID string, code string) error

// ValidateRecoveryCode validates and consumes a recovery code
func (s *Service) ValidateRecoveryCode(ctx context.Context, userID string, code string) error

// RegenerateRecoveryCodes generates new recovery codes (invalidates old ones)
func (s *Service) RegenerateRecoveryCodes(ctx context.Context, userID string) ([]string, error)

// Invalidate invalidates all OTPs for a user and purpose
func (s *Service) Invalidate(ctx context.Context, userID string, purpose Purpose) error

// Cleanup removes expired OTPs (for scheduled cleanup jobs)
func (s *Service) Cleanup(ctx context.Context) (int64, error)
```

### Request/Response Types

```go
// GenerateRequest contains parameters for OTP generation
type GenerateRequest struct {
    UserID    string
    Purpose   Purpose
    Channel   Channel            // Optional: delivery channel
    Recipient string             // Optional: where to send (email/phone)
    Metadata  map[string]string  // Optional: additional context
    Locale    string             // Optional: for localized messages
}

// GenerateResponse contains the result of OTP generation
type GenerateResponse struct {
    // OTP ID (for reference)
    ID string

    // The OTP code (only returned if notifier is not configured)
    Code string

    // Expiration time
    ExpiresAt time.Time

    // Whether the OTP was sent via notifier
    Sent bool

    // Remaining attempts
    RemainingAttempts int

    // When a new OTP can be requested
    ResendAvailableAt time.Time
}

// VerifyRequest contains parameters for OTP verification
type VerifyRequest struct {
    UserID  string
    Code    string
    Purpose Purpose
}

// VerifyResponse contains the result of OTP verification
type VerifyResponse struct {
    // Whether the code was valid
    Valid bool

    // Remaining attempts (if invalid)
    RemainingAttempts int

    // Whether the OTP is now locked out
    LockedOut bool

    // When lockout expires (if locked out)
    LockoutExpiresAt *time.Time

    // Error message (if invalid)
    Message string
}

// ResendRequest contains parameters for OTP resend
type ResendRequest struct {
    UserID    string
    Purpose   Purpose
    Channel   Channel
    Recipient string
}

// SetupTOTPRequest contains parameters for TOTP setup
type SetupTOTPRequest struct {
    UserID      string
    AccountName string  // Usually email
    Issuer      string  // Optional: override default issuer
}

// VerifyTOTPRequest contains parameters for TOTP verification
type VerifyTOTPRequest struct {
    UserID string
    Code   string
}
```

---

## Functional Options

```go
type Option func(*Service) error

// WithLogger sets a custom logger
func WithLogger(logger Logger) Option {
    return func(s *Service) error {
        if logger != nil {
            s.logger = logger
        }
        return nil
    }
}

// WithNotifier sets the notification provider
func WithNotifier(notifier Notifier) Option {
    return func(s *Service) error {
        s.notifier = notifier
        return nil
    }
}

// WithGenerator sets a custom OTP generator
func WithGenerator(generator Generator) Option {
    return func(s *Service) error {
        if generator != nil {
            s.generator = generator
        }
        return nil
    }
}

// WithAuditLogger sets the audit logger
func WithAuditLogger(auditLogger AuditLogger) Option {
    return func(s *Service) error {
        s.auditLogger = auditLogger
        return nil
    }
}

// WithCodeLength overrides the default code length
func WithCodeLength(length int) Option {
    return func(s *Service) error {
        if length < 4 || length > 10 {
            return errors.New("code length must be between 4 and 10")
        }
        s.config.CodeLength = length
        return nil
    }
}

// WithTTL overrides the default TTL
func WithTTL(ttl time.Duration) Option {
    return func(s *Service) error {
        if ttl < time.Minute {
            return errors.New("TTL must be at least 1 minute")
        }
        s.config.TTL = ttl
        return nil
    }
}

// WithMaxAttempts overrides the maximum verification attempts
func WithMaxAttempts(attempts int) Option {
    return func(s *Service) error {
        if attempts < 1 {
            return errors.New("max attempts must be at least 1")
        }
        s.config.MaxAttempts = attempts
        return nil
    }
}

// WithRateLimiting enables/disables rate limiting
func WithRateLimiting(enabled bool) Option {
    return func(s *Service) error {
        s.config.RateLimit.Enabled = enabled
        return nil
    }
}
```

---

## Error Types

```go
var (
    // ErrOTPNotFound indicates the OTP was not found
    ErrOTPNotFound = errors.New("otp not found")

    // ErrOTPExpired indicates the OTP has expired
    ErrOTPExpired = errors.New("otp has expired")

    // ErrOTPAlreadyUsed indicates the OTP has already been used
    ErrOTPAlreadyUsed = errors.New("otp has already been used")

    // ErrInvalidCode indicates the OTP code is invalid
    ErrInvalidCode = errors.New("invalid otp code")

    // ErrMaxAttemptsExceeded indicates too many failed verification attempts
    ErrMaxAttemptsExceeded = errors.New("maximum verification attempts exceeded")

    // ErrResendCooldown indicates a new OTP cannot be sent yet
    ErrResendCooldown = errors.New("please wait before requesting a new code")

    // ErrRateLimitExceeded indicates too many OTP requests
    ErrRateLimitExceeded = errors.New("too many otp requests")

    // ErrTOTPNotEnabled indicates TOTP is not enabled for the user
    ErrTOTPNotEnabled = errors.New("totp not enabled for user")

    // ErrTOTPAlreadyEnabled indicates TOTP is already enabled
    ErrTOTPAlreadyEnabled = errors.New("totp already enabled for user")

    // ErrInvalidRecoveryCode indicates the recovery code is invalid
    ErrInvalidRecoveryCode = errors.New("invalid recovery code")

    // ErrNoRecoveryCodesLeft indicates all recovery codes have been used
    ErrNoRecoveryCodesLeft = errors.New("no recovery codes remaining")

    // ErrNotificationFailed indicates OTP delivery failed
    ErrNotificationFailed = errors.New("failed to send otp notification")

    // ErrInvalidPurpose indicates an invalid OTP purpose
    ErrInvalidPurpose = errors.New("invalid otp purpose")
)

// OTPError provides structured error information
type OTPError struct {
    Code       string
    Message    string
    HTTPStatus int
    Retryable  bool
    RetryAfter time.Duration
}

func (e *OTPError) Error() string {
    return e.Message
}
```

---

## Logger Interface

```go
// Logger defines the logging contract (matches other packages)
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}

// NoopLogger is a logger that does nothing
type NoopLogger struct{}

func (l *NoopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *NoopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *NoopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *NoopLogger) Error(msg string, keysAndValues ...interface{}) {}
```

---

## Built-in Generators

```go
// DefaultGenerator provides built-in OTP generation algorithms
type DefaultGenerator struct {
    // Optional: custom random source for testing
    randSource io.Reader
}

// NewDefaultGenerator creates a new default generator
func NewDefaultGenerator() *DefaultGenerator {
    return &DefaultGenerator{
        randSource: rand.Reader,
    }
}

// Generate creates an OTP code based on the algorithm
func (g *DefaultGenerator) Generate(ctx context.Context, params *GenerateParams) (string, error) {
    switch params.Algorithm {
    case AlgorithmNumeric:
        return g.generateNumeric(params.Length)
    case AlgorithmTOTP:
        return g.generateTOTP(params.Secret, time.Now())
    case AlgorithmHOTP:
        return g.generateHOTP(params.Secret, params.Counter)
    default:
        return "", fmt.Errorf("unsupported algorithm: %s", params.Algorithm)
    }
}

// Validate checks if a code is valid
func (g *DefaultGenerator) Validate(ctx context.Context, params *ValidateParams) (bool, error) {
    switch params.Algorithm {
    case AlgorithmNumeric:
        // For numeric codes, validation is done by comparing hashes in the service
        return true, nil
    case AlgorithmTOTP:
        return g.validateTOTP(params.Code, params.Secret, params.Window)
    case AlgorithmHOTP:
        return g.validateHOTP(params.Code, params.Secret, params.Counter)
    default:
        return false, fmt.Errorf("unsupported algorithm: %s", params.Algorithm)
    }
}
```

---

## Example Implementations

### Storage: In-Memory Repository (for testing/development)

```go
// testutil/mocks.go

type InMemoryOTPRepository struct {
    mu   sync.RWMutex
    otps map[string]*otp.OTP
}

func NewInMemoryOTPRepository() *InMemoryOTPRepository {
    return &InMemoryOTPRepository{
        otps: make(map[string]*otp.OTP),
    }
}

func (r *InMemoryOTPRepository) Create(ctx context.Context, o *otp.OTP) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.otps[o.ID] = o
    return nil
}

// ... other methods
```

### Storage: PostgreSQL Repository

```go
// Can be provided as a separate package or example

type PostgresOTPRepository struct {
    db *postgres.Client
}

func NewPostgresOTPRepository(db *postgres.Client) *PostgresOTPRepository {
    return &PostgresOTPRepository{db: db}
}

func (r *PostgresOTPRepository) Create(ctx context.Context, o *otp.OTP) error {
    query := `
        INSERT INTO otps (id, user_id, code_hash, purpose, algorithm, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    _, err := r.db.Exec(ctx, query, o.ID, o.UserID, o.CodeHash, o.Purpose, o.Algorithm, o.ExpiresAt, o.CreatedAt)
    return err
}

// ... other methods
```

### Storage: Redis Repository

```go
// Example Redis implementation

type RedisOTPRepository struct {
    client *redis.Client
    prefix string
}

func (r *RedisOTPRepository) Create(ctx context.Context, o *otp.OTP) error {
    data, _ := json.Marshal(o)
    key := fmt.Sprintf("%s:%s", r.prefix, o.ID)
    return r.client.Set(ctx, key, data, time.Until(o.ExpiresAt)).Err()
}

// ... other methods
```

### Notifier: SMS (Twilio)

```go
// Example Twilio implementation

type TwilioNotifier struct {
    client    *twilio.Client
    fromPhone string
    templates map[otp.Purpose]string
}

func (n *TwilioNotifier) Send(ctx context.Context, req *otp.NotificationRequest) error {
    template := n.templates[req.Purpose]
    message := fmt.Sprintf(template, req.Code)

    _, err := n.client.Messages.Create(ctx, &twilio.CreateMessageParams{
        To:   req.Recipient,
        From: n.fromPhone,
        Body: message,
    })
    return err
}
```

### Notifier: Email (SendGrid)

```go
// Example SendGrid implementation

type SendGridNotifier struct {
    client    *sendgrid.Client
    fromEmail string
    templates map[otp.Purpose]string
}

func (n *SendGridNotifier) Send(ctx context.Context, req *otp.NotificationRequest) error {
    templateID := n.templates[req.Purpose]

    message := mail.NewSingleEmail(
        mail.NewEmail("MyApp", n.fromEmail),
        "Your verification code",
        mail.NewEmail("", req.Recipient),
        fmt.Sprintf("Your code is: %s", req.Code),
        "",
    )
    message.SetTemplateID(templateID)
    message.AddDynamicTemplateData("code", req.Code)

    _, err := n.client.Send(message)
    return err
}
```

### Notifier: Multi-Channel

```go
// Example multi-channel notifier

type MultiChannelNotifier struct {
    sms   otp.Notifier
    email otp.Notifier
    push  otp.Notifier
}

func (n *MultiChannelNotifier) Send(ctx context.Context, req *otp.NotificationRequest) error {
    switch req.Channel {
    case otp.ChannelSMS:
        return n.sms.Send(ctx, req)
    case otp.ChannelEmail:
        return n.email.Send(ctx, req)
    case otp.ChannelPush:
        return n.push.Send(ctx, req)
    case otp.ChannelAuto:
        // Implement fallback logic
        if err := n.sms.Send(ctx, req); err == nil {
            return nil
        }
        return n.email.Send(ctx, req)
    default:
        return fmt.Errorf("unsupported channel: %s", req.Channel)
    }
}
```

---

## Integration with Auth Package

The OTP package can be integrated with the existing `auth` package:

```go
// In auth package, add OTP support

type OTPService interface {
    Generate(ctx context.Context, req *otp.GenerateRequest) (*otp.GenerateResponse, error)
    Verify(ctx context.Context, req *otp.VerifyRequest) (*otp.VerifyResponse, error)
}

// AuthService can optionally use OTP for 2FA
type Service struct {
    // ... existing fields
    otpService OTPService // Optional
}

// Login with 2FA support
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    user, err := s.validateCredentials(ctx, email, password)
    if err != nil {
        return nil, err
    }

    // Check if 2FA is enabled
    if user.TwoFactorEnabled && s.otpService != nil {
        // Generate and send OTP
        _, err := s.otpService.Generate(ctx, &otp.GenerateRequest{
            UserID:  user.ID,
            Purpose: otp.PurposeTwoFactor,
        })
        if err != nil {
            return nil, err
        }

        return &LoginResult{
            RequiresTwoFactor: true,
            TwoFactorToken:    generateTempToken(user.ID),
        }, nil
    }

    return s.createSession(ctx, user)
}

// VerifyTwoFactor completes login with OTP
func (s *Service) VerifyTwoFactor(ctx context.Context, tempToken, code string) (*LoginResult, error) {
    userID := validateTempToken(tempToken)

    result, err := s.otpService.Verify(ctx, &otp.VerifyRequest{
        UserID:  userID,
        Code:    code,
        Purpose: otp.PurposeTwoFactor,
    })
    if err != nil {
        return nil, err
    }
    if !result.Valid {
        return nil, ErrInvalidCode
    }

    user, _ := s.repos.Users.GetByID(ctx, userID)
    return s.createSession(ctx, user)
}
```

---

## HTTP Middleware Helpers

```go
// middleware.go

// ExtractOTPFromRequest extracts OTP from various sources
func ExtractOTPFromRequest(r *http.Request) string {
    // Check header first
    if code := r.Header.Get("X-OTP-Code"); code != "" {
        return code
    }

    // Check query parameter
    if code := r.URL.Query().Get("otp"); code != "" {
        return code
    }

    // Check form value
    if code := r.FormValue("otp"); code != "" {
        return code
    }

    return ""
}

// RequireOTP creates middleware that requires OTP verification
func RequireOTP(otpService *Service, purpose Purpose) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := getUserIDFromContext(r.Context())
            code := ExtractOTPFromRequest(r)

            if code == "" {
                http.Error(w, "OTP code required", http.StatusUnauthorized)
                return
            }

            result, err := otpService.Verify(r.Context(), &VerifyRequest{
                UserID:  userID,
                Code:    code,
                Purpose: purpose,
            })
            if err != nil || !result.Valid {
                http.Error(w, "Invalid OTP code", http.StatusUnauthorized)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Usage Examples

### Basic OTP Flow

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/example/core-backend/pkg/otp"
    "github.com/example/core-backend/pkg/otp/testutil"
)

func main() {
    // Create in-memory repository (use PostgresOTPRepository in production)
    repo := testutil.NewInMemoryOTPRepository()

    // Create service with default configuration
    cfg := otp.DefaultConfig()
    svc, err := otp.NewService(cfg, repo)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // Generate OTP
    genResp, err := svc.Generate(ctx, &otp.GenerateRequest{
        UserID:  "user-123",
        Purpose: otp.PurposeEmailVerify,
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("OTP Code: %s (expires at %s)\n", genResp.Code, genResp.ExpiresAt)

    // Verify OTP
    verifyResp, err := svc.Verify(ctx, &otp.VerifyRequest{
        UserID:  "user-123",
        Code:    genResp.Code,
        Purpose: otp.PurposeEmailVerify,
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Valid: %v\n", verifyResp.Valid)
}
```

### With SMS Notification

```go
package main

import (
    "github.com/example/core-backend/pkg/otp"
)

func main() {
    repo := NewPostgresOTPRepository(db)
    twilioNotifier := NewTwilioNotifier(twilioClient, "+1234567890")

    svc, _ := otp.NewService(
        otp.DefaultConfig(),
        repo,
        otp.WithNotifier(twilioNotifier),
        otp.WithLogger(logger),
        otp.WithTTL(10 * time.Minute),
    )

    // Generate and send OTP via SMS
    _, err := svc.Generate(ctx, &otp.GenerateRequest{
        UserID:    "user-123",
        Purpose:   otp.PurposeLogin,
        Channel:   otp.ChannelSMS,
        Recipient: "+1987654321",
    })
    // OTP is automatically sent via Twilio
}
```

### TOTP Setup (Authenticator Apps)

```go
package main

import (
    "github.com/example/core-backend/pkg/otp"
)

func main() {
    svc, _ := otp.NewService(cfg, repo)

    // Setup TOTP for user
    setup, err := svc.SetupTOTP(ctx, &otp.SetupTOTPRequest{
        UserID:      "user-123",
        AccountName: "user@example.com",
    })
    if err != nil {
        panic(err)
    }

    // Display QR code to user
    fmt.Printf("Scan this QR code: %s\n", setup.QRCodeURL)
    fmt.Printf("Or enter manually: %s\n", setup.ManualEntryKey)
    fmt.Printf("Recovery codes: %v\n", setup.RecoveryCodes)

    // User scans QR code and enters the code from their app
    code := "123456" // From authenticator app

    // Enable TOTP after verification
    err = svc.EnableTOTP(ctx, "user-123", code)
    if err != nil {
        panic(err)
    }

    // Later, verify TOTP during login
    result, _ := svc.VerifyTOTP(ctx, &otp.VerifyTOTPRequest{
        UserID: "user-123",
        Code:   "654321",
    })
    fmt.Printf("TOTP valid: %v\n", result.Valid)
}
```

---

## Database Schema (Reference)

```sql
-- For PostgreSQL implementations

CREATE TABLE otps (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    code_hash VARCHAR(64) NOT NULL,
    purpose VARCHAR(32) NOT NULL,
    algorithm VARCHAR(16) NOT NULL DEFAULT 'numeric',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

CREATE INDEX idx_otps_user_purpose ON otps(user_id, purpose);
CREATE INDEX idx_otps_expires_at ON otps(expires_at);

-- For TOTP secrets
CREATE TABLE totp_secrets (
    user_id VARCHAR(36) PRIMARY KEY,
    secret_encrypted BYTEA NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    enabled_at TIMESTAMP WITH TIME ZONE
);

-- For recovery codes
CREATE TABLE recovery_codes (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    code_hash VARCHAR(64) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_recovery_codes_user ON recovery_codes(user_id);
```

---

## Security Considerations

1. **Code Hashing**: OTP codes are hashed (SHA-256) before storage to prevent database leaks from exposing valid codes.

2. **Rate Limiting**: Built-in rate limiting prevents brute-force attacks on OTP verification.

3. **Attempt Limiting**: Maximum verification attempts prevent unlimited guessing.

4. **Secure Generation**: Uses `crypto/rand` for cryptographically secure random number generation.

5. **TOTP Secret Encryption**: TOTP secrets should be encrypted at rest (implementation responsibility).

6. **Timing-Safe Comparison**: Code verification uses constant-time comparison to prevent timing attacks.

7. **Short TTL**: Default 5-minute expiration limits the window for attacks.

8. **Audit Logging**: Optional audit logging tracks all OTP-related events.

---

## Implementation Phases

### Phase 1: Core Implementation
- [ ] Basic OTP model and repository interface
- [ ] Default numeric code generator
- [ ] Service with Generate/Verify methods
- [ ] Configuration with environment loading
- [ ] Error types and logging

### Phase 2: Notification Support
- [ ] Notifier interface
- [ ] Multi-channel support
- [ ] Resend with cooldown

### Phase 3: TOTP Support
- [ ] TOTP generation (RFC 6238)
- [ ] HOTP generation (RFC 4226)
- [ ] QR code URL generation
- [ ] Recovery codes

### Phase 4: Testing & Examples
- [ ] In-memory repository mock
- [ ] Unit tests
- [ ] Integration examples
- [ ] PostgreSQL repository example
- [ ] Redis repository example
- [ ] Twilio/SendGrid notifier examples

### Phase 5: Integration
- [ ] HTTP middleware helpers
- [ ] Auth package integration
- [ ] gRPC interceptor helpers

---

## Summary

This OTP package design follows the established patterns in the core-backend library:

| Pattern | Implementation |
|---------|----------------|
| Interface-based DI | `OTPRepository`, `Notifier`, `Generator`, `AuditLogger` |
| Functional Options | `WithLogger`, `WithNotifier`, `WithGenerator`, etc. |
| Environment Config | `LoadConfig()` reads `OTP_*` environment variables |
| Sensible Defaults | `DefaultConfig()` provides production-ready defaults |
| Minimal Logger | Same 4-method interface as other packages |
| Package-specific Errors | Sentinel errors + structured `OTPError` type |
| Testing Support | `testutil/mocks.go` with in-memory implementations |
| Examples | Separate examples for each use case |

The package is completely agnostic of:
- **Storage**: Any database via `OTPRepository` interface
- **Notification**: Any delivery method via `Notifier` interface
- **Algorithm**: Pluggable via `Generator` interface
- **Logging**: Any logger via `Logger` interface
