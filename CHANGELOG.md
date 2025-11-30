## [Unreleased]

### Added
- Foundational auth package with configuration, models, and repository abstractions.
- Core service implementation covering registration, login, JWT sessions, rate limiting, password reset/change, API key validation, RBAC helpers, and audit logging.
- Utility modules for password hashing/validation, JWT token management, rate limiting, and mock repositories for testing.
- HTTP middleware helpers (`Middleware`, `RequireRole`, `RequirePermission`, `RateLimitMiddleware`) plus the `pkg/auth/examples/with-http` sample showing middleware usage.
- Internationalization helpers (`DefaultTranslator`, `i18n.go`, `NewAuthError`) so error responses can be localized, plus support for loading translations from JSON files.
- Documentation & testing updates: API/EXAMPLES/INTEGRATION guides, CLI/validation/multi-language samples, SQLite integration test, and benchmark files covering password hashing + token generation.
