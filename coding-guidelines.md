# Go Coding Guidelines

> **Role**: You are a **Senior Go Software Engineer** with deep expertise in writing production-grade, maintainable, and high-performance Go code. Apply industry best practices, idiomatic Go patterns, and software engineering excellence in all implementations.

## Core Philosophy

As a senior engineer, you must:
- **Write code that others will maintain**: Prioritize clarity, simplicity, and maintainability
- **Follow Go idioms**: Embrace "The Go Way" - simplicity, explicit error handling, composition over inheritance
- **Think in systems**: Consider scalability, performance, observability, and operational excellence
- **Default to standard library**: Only introduce dependencies when they provide significant value
- **Optimize for readability**: Code is read far more often than it's written
- **Build with production in mind**: Security, error handling, logging, metrics, graceful degradation

## Best Practices Mindset

**Senior engineers ask:**
1. Is this the simplest solution that could work?
2. Have I checked for existing implementations before writing new code?
3. Will this code be easy to test, debug, and modify?
4. What failure modes exist, and have I handled them?
5. Is this code performant enough for production load?
6. Does this follow Go community conventions and idioms?

## Design Principles

### SOLID Principles

**S - Single Responsibility Principle (SRP)**
- Each type/function should have one reason to change
- Separate concerns into distinct packages and types
- Example: Split data access, business logic, and API handling into different layers

```go
// Bad: UserService handles everything
type UserService struct {
    db *sql.DB
}
func (s *UserService) CreateUser(user User) error {
    // validation, business logic, database access all mixed
}

// Good: Separated concerns
type UserValidator struct {}
func (v *UserValidator) Validate(user User) error { /* validation only */ }

type UserRepository struct { db *sql.DB }
func (r *UserRepository) Save(user User) error { /* data access only */ }

type UserService struct {
    validator  *UserValidator
    repository *UserRepository
}
func (s *UserService) CreateUser(user User) error {
    if err := s.validator.Validate(user); err != nil {
        return err
    }
    return s.repository.Save(user)
}
```

**O - Open/Closed Principle (OCP)**
- Open for extension, closed for modification
- Use interfaces and composition over direct implementation

```go
// Extensible through interfaces
type Notifier interface {
    Notify(ctx context.Context, message string) error
}

type NotificationService struct {
    notifiers []Notifier
}

// Add new notifiers without modifying existing code
func (s *NotificationService) AddNotifier(n Notifier) {
    s.notifiers = append(s.notifiers, n)
}
```

**L - Liskov Substitution Principle (LSP)**
- Subtypes must be substitutable for their base types
- Interface implementations should honor contracts

```go
type Storage interface {
    Save(ctx context.Context, key string, value []byte) error
    Load(ctx context.Context, key string) ([]byte, error)
}

// Both implementations must behave consistently
type FileStorage struct {}
type S3Storage struct {}
```

**I - Interface Segregation Principle (ISP)**
- Many small, specific interfaces over one large interface
- Clients shouldn't depend on methods they don't use

```go
// Bad: Fat interface
type Repository interface {
    Create(user User) error
    Read(id string) (User, error)
    Update(user User) error
    Delete(id string) error
    Search(query string) ([]User, error)
    Export() ([]byte, error)
}

// Good: Segregated interfaces
type UserReader interface {
    Read(id string) (User, error)
}

type UserWriter interface {
    Create(user User) error
    Update(user User) error
}

type UserDeleter interface {
    Delete(id string) error
}
```

**D - Dependency Inversion Principle (DIP)**
- Depend on abstractions, not concretions
- Use dependency injection

```go
// Good: Depends on abstraction
type OrderService struct {
    repo   OrderRepository  // interface
    notify Notifier        // interface
}

func NewOrderService(repo OrderRepository, notify Notifier) *OrderService {
    return &OrderService{
        repo:   repo,
        notify: notify,
    }
}
```

### DRY Principle (Don't Repeat Yourself)

**Before creating new functions:**
1. Search for existing implementations in the codebase
2. Check if existing functions can be reused or extended
3. Extract common patterns into shared utilities
4. Use composition to build complex behavior from simple functions

```go
// Bad: Duplicated validation logic
func CreateUser(user User) error {
    if user.Email == "" || !strings.Contains(user.Email, "@") {
        return errors.New("invalid email")
    }
    // ...
}

func UpdateUser(user User) error {
    if user.Email == "" || !strings.Contains(user.Email, "@") {
        return errors.New("invalid email")
    }
    // ...
}

// Good: Reuse validation
func validateEmail(email string) error {
    if email == "" || !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

func CreateUser(user User) error {
    if err := validateEmail(user.Email); err != nil {
        return err
    }
    // ...
}

func UpdateUser(user User) error {
    if err := validateEmail(user.Email); err != nil {
        return err
    }
    // ...
}
```

**Code Reuse Strategies:**

```go
// 1. Extract common logic into helper functions
func parseAndValidateID(idStr string) (int64, error) {
    if idStr == "" {
        return 0, errors.New("id is required")
    }
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid id format: %w", err)
    }
    if id <= 0 {
        return 0, errors.New("id must be positive")
    }
    return id, nil
}

// 2. Use function composition for complex operations
func processUser(user User) error {
    validators := []func(User) error{
        validateEmail,
        validateAge,
        validateName,
    }

    for _, validate := range validators {
        if err := validate(user); err != nil {
            return err
        }
    }
    return nil
}

// 3. Build on existing functions rather than duplicating
func validateAndSaveUser(user User) error {
    if err := processUser(user); err != nil {  // Reuse existing validation
        return err
    }
    return saveToDatabase(user)  // Reuse existing save logic
}
```

**When to Create New Functions:**
- Only create new functions when existing ones cannot be reasonably reused
- If you need to modify copied code more than 20%, extract a new abstraction
- Always prefer composition and parameterization over duplication

## Code Organization

- **Package naming**: Use short, lowercase, single-word names
- **File naming**: Use snake_case (e.g., `user_service.go`)
- **One concept per file**: Group related functions together
- **Keep files under 500 lines** when possible

## Naming Conventions

```go
// Exported (public)
type UserService struct {}
func NewUserService() *UserService {}
const MaxRetries = 3

// Unexported (private)
type userCache struct {}
func validateInput() error {}
const defaultTimeout = 30
```

## Error Handling

```go
// Always check errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use custom errors for domain logic
var ErrUserNotFound = errors.New("user not found")

// Wrap errors with context
if err != nil {
    return fmt.Errorf("processing user %s: %w", userID, err)
}
```

## Function Design

- **Keep functions small**: Aim for under 50 lines
- **Single responsibility**: One function, one purpose
- **Early returns**: Reduce nesting with guard clauses

```go
func ProcessUser(id string) error {
    if id == "" {
        return ErrInvalidInput
    }

    user, err := fetchUser(id)
    if err != nil {
        return err
    }

    return processData(user)
}
```

## Testing

- **Test file naming**: `*_test.go`
- **Test function naming**: `TestFunctionName_Scenario`
- **Table-driven tests** for multiple cases
- **Target ≥80% coverage**

```go
func TestValidateEmail_ValidFormats(t *testing.T) {
    tests := []struct {
        name  string
        email string
        want  bool
    }{
        {"valid email", "user@example.com", true},
        {"missing @", "userexample.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Enums

Go doesn't have native enums, but use typed constants with `iota`:

```go
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusInactive
    StatusDeleted
)

// String method for readable output
func (s Status) String() string {
    switch s {
    case StatusPending:
        return "pending"
    case StatusActive:
        return "active"
    case StatusInactive:
        return "inactive"
    case StatusDeleted:
        return "deleted"
    default:
        return "unknown"
    }
}

// ParseStatus for string-to-enum conversion
func ParseStatus(s string) (Status, error) {
    switch strings.ToLower(s) {
    case "pending":
        return StatusPending, nil
    case "active":
        return StatusActive, nil
    case "inactive":
        return StatusInactive, nil
    case "deleted":
        return StatusDeleted, nil
    default:
        return 0, fmt.Errorf("invalid status: %s", s)
    }
}
```

## Configuration

- **Use environment variables** for deployment-specific settings
- **Provide defaults** for all configuration values
- **Validate on startup** to fail fast
- **Centralize configuration** in a config package

```go
type Config struct {
    ServerPort    int           `env:"SERVER_PORT" envDefault:"8080"`
    DatabaseURL   string        `env:"DATABASE_URL,required"`
    LogLevel      string        `env:"LOG_LEVEL" envDefault:"info"`
    Timeout       time.Duration `env:"TIMEOUT" envDefault:"30s"`
    MaxRetries    int           `env:"MAX_RETRIES" envDefault:"3"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}

    // Using env parsing library (e.g., caarlos0/env)
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    return cfg, nil
}

func (c *Config) Validate() error {
    if c.ServerPort < 1 || c.ServerPort > 65535 {
        return errors.New("invalid server port")
    }
    if c.Timeout < 1*time.Second {
        return errors.New("timeout too short")
    }
    return nil
}
```

### Alternative: YAML/JSON Config Files

```go
type Config struct {
    Server   ServerConfig   `yaml:"server" json:"server"`
    Database DatabaseConfig `yaml:"database" json:"database"`
}

type ServerConfig struct {
    Port    int    `yaml:"port" json:"port"`
    Host    string `yaml:"host" json:"host"`
}

func LoadConfigFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }

    cfg := &Config{}
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }

    return cfg, nil
}
```

### Hybrid Approach (Recommended)

Load defaults from file, override with environment variables:

```go
func LoadConfig() (*Config, error) {
    // Load from file with defaults
    cfg, err := loadFromFile("config.yaml")
    if err != nil {
        cfg = defaultConfig()
    }

    // Override with environment variables
    if port := os.Getenv("SERVER_PORT"); port != "" {
        cfg.Server.Port, _ = strconv.Atoi(port)
    }
    if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
        cfg.Database.URL = dbURL
    }

    return cfg, cfg.Validate()
}
```

## Interfaces

- **Small interfaces**: Prefer 1-3 methods
- **Accept interfaces, return structs**
- **Name with -er suffix**: `Reader`, `Writer`, `Handler`

```go
type UserFetcher interface {
    FetchUser(ctx context.Context, id string) (*User, error)
}
```

## Context Usage

- **First parameter**: Always `ctx context.Context`
- **Never store**: Pass context through function calls
- **Respect cancellation**: Check `ctx.Done()` in long operations

```go
func ProcessData(ctx context.Context, data []byte) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // process data
    }
}
```

## Concurrency

- **Always use context** for cancellation
- **Protect shared state** with mutexes or channels
- **Wait for goroutines** with `sync.WaitGroup` or `errgroup`

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item string) {
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()
```

## Comments

- **Exported symbols**: Must have doc comments
- **Start with the name**: "ProcessUser processes..."
- **Explain why, not what**: Code shows what, comments explain why

```go
// ProcessUser validates and processes user data.
// Returns ErrInvalidInput if validation fails.
func ProcessUser(user *User) error {
    // Normalize email to lowercase for case-insensitive comparison
    user.Email = strings.ToLower(user.Email)
    return validate(user)
}
```

## Common Patterns

### Initialization
```go
func NewService(db *sql.DB, logger *log.Logger) *Service {
    return &Service{
        db:     db,
        logger: logger,
    }
}
```

### Options Pattern
```go
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}

func NewClient(opts ...Option) *Client {
    cfg := &Config{Timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(cfg)
    }
    return &Client{config: cfg}
}
```

### Cleanup
```go
func Open() (*Resource, error) {
    r := &Resource{}
    if err := r.init(); err != nil {
        return nil, err
    }
    return r, nil
}

func (r *Resource) Close() error {
    // cleanup
    return nil
}

// Usage
resource, err := Open()
if err != nil {
    return err
}
defer resource.Close()
```

## Code Quality

- Run `golangci-lint run` before committing
- Use `gofmt` for formatting (automatic in most editors)
- Keep cyclomatic complexity low (max 10 per function)
- Avoid global mutable state

## Dependencies

- Use standard library when possible
- Keep `go.mod` clean and up-to-date
- Run `go mod tidy` regularly
- Vendor if needed for stability

## Performance

- **Profile before optimizing**
- **Reuse buffers and connections**
- **Use sync.Pool for frequent allocations**
- **Avoid premature optimization**

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

## Security

- **Validate all input**
- **Use context timeouts** to prevent resource exhaustion
- **Never log sensitive data** (passwords, tokens, PII)
- **Use prepared statements** for SQL queries

## API Response Format

All API responses should follow a consistent structure to improve client-side handling and debugging.

### Mandatory Request Headers

All API requests **MUST** include the following headers for proper request handling, tracing, and analytics:

#### X-Timezone (mandatory)
```
X-Timezone: Asia/Jakarta
```

- **Required**: Yes (for all API requests)
- **Format**: IANA timezone identifier (e.g., `Asia/Jakarta`, `America/New_York`, `Europe/London`, `UTC`)
- **Purpose**:
  - Interprets date-only parameters (YYYY-MM-DD) into datetime ranges in client's timezone
  - Determines day/week/month boundaries for statistics and grouping
  - Ensures consistent date handling across different client timezones
- **Validation**: Must be a valid IANA timezone identifier from the tz database
- **Error if missing**: 400 Bad Request with error message

**Example:**
```go
// Middleware to validate X-Timezone header
func ValidateTimezoneHeader(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        timezone := c.Request().Header.Get("X-Timezone")
        if timezone == "" {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Status:  false,
                Error:   "validation_error",
                Message: "X-Timezone header is required",
            })
        }

        // Validate timezone
        if _, err := time.LoadLocation(timezone); err != nil {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Status:  false,
                Error:   "validation_error",
                Message: "Invalid timezone identifier",
            })
        }

        return next(c)
    }
}
```

#### X-Request-ID (mandatory)
```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

- **Required**: Yes (for all API requests)
- **Format**: UUID v4 or any unique string (max 128 characters)
- **Purpose**:
  - Distributed tracing across services
  - Request correlation in logs
  - Debugging and troubleshooting
  - Performance monitoring
- **Server behavior**: If not provided by client, server will generate one automatically
- **Response**: Server returns the same X-Request-ID in response headers for correlation

**Example:**
```go
// Middleware to ensure X-Request-ID exists
func EnsureRequestID(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        requestID := c.Request().Header.Get("X-Request-ID")

        // Generate if not provided
        if requestID == "" {
            requestID = uuid.New().String()
            c.Request().Header.Set("X-Request-ID", requestID)
        }

        // Add to response headers
        c.Response().Header().Set("X-Request-ID", requestID)

        // Add to context for logging
        c.Set("request_id", requestID)

        return next(c)
    }
}
```

#### X-Request-Origin (mandatory)
```
X-Request-Origin: mobile-android
```

- **Required**: Yes (for all API requests)
- **Format**: Free text string (max 50 characters)
- **Purpose**:
  - Track request sources for analytics
  - Monitor API usage by platform
  - Debug platform-specific issues
  - Rate limiting per platform
- **Common values**: `web-app`, `mobile-ios`, `mobile-android`, `telegram-bot`, `api-client`
- **Validation**: Alphanumeric characters and hyphens only
- **Default**: If not provided, server sets to `unknown` (but client should always send)

**Example:**
```go
// Middleware to validate and track request origin
func ValidateRequestOrigin(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        origin := c.Request().Header.Get("X-Request-Origin")

        // Set default if missing (but log warning)
        if origin == "" {
            origin = "unknown"
            c.Logger().Warnf("Missing X-Request-Origin header from %s", c.RealIP())
        }

        // Validate format (alphanumeric and hyphens only, max 50 chars)
        if len(origin) > 50 || !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(origin) {
            return c.JSON(http.StatusBadRequest, ErrorResponse{
                Status:  false,
                Error:   "validation_error",
                Message: "Invalid X-Request-Origin format",
            })
        }

        // Add to context for logging
        c.Set("request_origin", origin)

        return next(c)
    }
}
```

### Header Validation Order

Validate headers in this order:
1. **X-Request-ID**: Generate if missing, validate format
2. **X-Timezone**: Validate presence and format (required)
3. **X-Request-Origin**: Validate format, default to "unknown" if missing

### Complete Handler Example

```go
func (h *Handler) GetTransactions(c echo.Context) error {
    // Headers are already validated by middleware
    requestID := c.Get("request_id").(string)
    timezone := c.Request().Header.Get("X-Timezone")
    origin := c.Get("request_origin").(string)

    // Log request with headers
    c.Logger().Infof("Request: id=%s, timezone=%s, origin=%s, path=%s",
        requestID, timezone, origin, c.Path())

    // Use timezone for date parsing
    location, _ := time.LoadLocation(timezone)

    // Process request...

    // Return success response
    return c.JSON(http.StatusOK, SuccessResponse{
        Status: true,
        Data:   transactions,
    })
}
```

### Success Response

```go
type SuccessResponse struct {
    Status bool        `json:"status"`
    Data   interface{} `json:"data"`
}

// Example
{
  "status": true,
  "data": {
    "id": "123",
    "name": "John Doe"
  }
}
```

### Error Response

All error responses must include a `status` field set to `false` and follow this structure:

```go
type ErrorResponse struct {
    Status  bool              `json:"status"`
    Error   string            `json:"error"`
    Message string            `json:"message"`
    Details []ValidationError `json:"details,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

**Example - Validation Error (400 Bad Request):**
```json
{
  "status": false,
  "error": "validation_error",
  "message": "Invalid input",
  "details": [
    {
      "field": "start_date",
      "message": "start_date must be before end_date"
    },
    {
      "field": "email",
      "message": "email must be a valid email address"
    }
  ]
}
```

**Example - Not Found Error (404 Not Found):**
```json
{
  "status": false,
  "error": "not_found",
  "message": "User not found"
}
```

**Example - Internal Server Error (500):**
```json
{
  "status": false,
  "error": "internal_server_error",
  "message": "An internal error occurred. Please try again later."
}
```

### Standard Error Codes

Use consistent error codes across the API:

- `validation_error` - Input validation failed (400)
- `bad_request` - Malformed request (400)
- `unauthorized` - Authentication required or failed (401)
- `forbidden` - Insufficient permissions (403)
- `not_found` - Resource not found (404)
- `conflict` - Resource conflict (409)
- `rate_limit_exceeded` - Too many requests (429)
- `internal_server_error` - Server-side error (500)
- `service_unavailable` - Service temporarily unavailable (503)
- `gateway_timeout` - Request timeout (504)

### Implementation Example

```go
// Helper function for error responses
func sendError(c echo.Context, statusCode int, errorCode, message string, details []ValidationError) error {
    return c.JSON(statusCode, ErrorResponse{
        Status:  false,
        Error:   errorCode,
        Message: message,
        Details: details,
    })
}

// Usage in handler
func (h *Handler) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return sendError(c, http.StatusBadRequest, "bad_request",
            "Malformed JSON in request body", nil)
    }

    validationErrors := req.Validate()
    if len(validationErrors) > 0 {
        return sendError(c, http.StatusBadRequest, "validation_error",
            "Invalid input", validationErrors)
    }

    // Success response
    return c.JSON(http.StatusCreated, SuccessResponse{
        Status: true,
        Data:   user,
    })
}
```

### API Request/Response Summary

**All API requests MUST include these headers:**
```bash
X-Timezone: Asia/Jakarta          # Mandatory - IANA timezone identifier
X-Request-ID: uuid-or-unique-id   # Mandatory - Generated if not provided
X-Request-Origin: mobile-android  # Mandatory - Request source identifier
```

**All API responses MUST:**
- Include `status` field (true for success, false for errors)
- Return `X-Request-ID` in response headers for correlation
- Use consistent error codes and structure
- Wrap success data in `SuccessResponse` with `status: true`

**Complete Request Example:**
```bash
curl -X GET "http://localhost:8080/api/v1/transactions?start_date=2025-01-01&end_date=2025-01-31" \
  -H "Content-Type: application/json" \
  -H "X-Timezone: Asia/Jakarta" \
  -H "X-Request-Origin: mobile-android" \
  -H "X-Request-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Complete Success Response Example:**
```json
{
  "status": true,
  "data": {
    "transactions": [...]
  }
}
```
**Response Headers:**
```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json
```

**Complete Error Response Example:**
```json
{
  "status": false,
  "error": "validation_error",
  "message": "Invalid input",
  "details": [
    {
      "field": "start_date",
      "message": "Invalid date format, use YYYY-MM-DD"
    }
  ]
}
```

### CORS Configuration

Ensure these headers are allowed in CORS configuration:

```go
// Echo CORS middleware configuration
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"http://localhost:3000", "https://yourdomain.com"},
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{
        echo.HeaderContentType,
        echo.HeaderAuthorization,
        "X-Timezone",         // Required for date handling
        "X-Request-ID",       // Required for tracing
        "X-Request-Origin",   // Required for analytics
    },
    ExposeHeaders: []string{
        "X-Request-ID",       // Expose for client-side correlation
    },
}))
```

---

## Senior Engineer Best Practices Summary

### Code Quality Checklist
- ✅ Follows SOLID principles and DRY methodology
- ✅ Reuses existing functions before creating new ones
- ✅ Uses idiomatic Go patterns (interfaces, composition, error handling)
- ✅ Keeps functions under 50 lines with single responsibility
- ✅ Has comprehensive table-driven tests (≥80% coverage)
- ✅ Includes doc comments for all exported symbols
- ✅ Handles all errors with proper context wrapping
- ✅ Uses appropriate concurrency patterns with context cancellation
- ✅ Optimizes for readability and maintainability
- ✅ Considers production concerns (security, performance, observability)

### API Request/Response Checklist
- ✅ All API handlers validate mandatory headers (X-Timezone, X-Request-ID, X-Request-Origin)
- ✅ All responses include `status` field (true/false)
- ✅ Success responses wrapped in `SuccessResponse{Status: true, Data: ...}`
- ✅ Error responses include error code, message, and optional details array
- ✅ X-Request-ID returned in response headers for correlation
- ✅ Timezone handling for date-only parameters (YYYY-MM-DD format)
- ✅ CORS configured to allow mandatory headers

### Go Idioms to Follow
```go
// Accept interfaces, return concrete types
func NewService(repo Repository) *Service { }

// Use functional options for extensibility
func NewClient(opts ...Option) *Client { }

// Prefer table-driven tests
func TestFunc(t *testing.T) { tests := []struct{...}{...} }

// Make the zero value useful
type Config struct { Timeout time.Duration } // defaults to 0

// Handle errors explicitly, never ignore them
if err != nil { return fmt.Errorf("context: %w", err) }

// Use defer for cleanup
defer file.Close()

// Keep interfaces small (1-3 methods preferred)
type Reader interface { Read(p []byte) (n int, err error) }
```

### Production Excellence
- **Observability**: Log appropriately (structured logging with levels)
- **Metrics**: Instrument critical paths for monitoring
- **Graceful degradation**: Handle partial failures elegantly
- **Resource management**: Close connections, cancel contexts, use timeouts
- **Backward compatibility**: Consider API versioning and migrations
- **Performance**: Profile before optimizing, use benchmarks to validate

**Remember**: As a senior engineer, your code should be exemplary - clean, tested, documented, and production-ready.
