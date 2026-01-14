## [Unreleased]

### Added
- **HTTP Client Package (`pkg/httpclient`)**: Production-grade HTTP client library with enterprise features
  - Fluent API for building HTTP requests (GET, POST, PUT, PATCH, DELETE)
  - Automatic retry with exponential backoff for transient failures (5xx errors, network issues)
  - Circuit breaker pattern implementation to prevent cascading failures
  - Middleware system for request/response interception (logging, auth, custom headers)
  - JSON encoding/decoding helpers
  - Context-aware operations with cancellation and timeout support
  - Comprehensive test suite with 80%+ coverage
  - Zero external dependencies (uses only Go standard library)
  - Complete documentation with examples and usage patterns
  - Package location: `github.com/rompi/core-backend/pkg/httpclient`
- Foundational auth package with configuration, models, and repository abstractions.
- Core service implementation covering registration, login, JWT sessions, rate limiting, password reset/change, API key validation, RBAC helpers, and audit logging.
- Utility modules for password hashing/validation, JWT token management, rate limiting, and mock repositories for testing.
- HTTP middleware helpers (`Middleware`, `RequireRole`, `RequirePermission`, `RateLimitMiddleware`) plus the `pkg/auth/examples/with-http` sample showing middleware usage.
- Internationalization helpers (`DefaultTranslator`, `i18n.go`, `NewAuthError`) so error responses can be localized, plus support for loading translations from JSON files.
- Documentation & testing updates: API/EXAMPLES/INTEGRATION guides, CLI/validation/multi-language samples, SQLite integration test, and benchmark files covering password hashing + token generation.
