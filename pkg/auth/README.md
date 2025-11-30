# Auth Package

The `pkg/auth` module provides a standalone authentication service that you can embed in any HTTP API or gRPC backend. It bundles secure password policies, JWT session handling, account lockout, role/permission helpers, API key validation, rate limiting, audit logging, and localization so you can focus on wiring persistence and transport logic.

## Design Snapshot

- **Service contract:** `Service` exposes registration, login, password reset, API-key validation, token refresh, role/permission checks, and helper middleware (`Middleware`, `RequireRole`, `RequirePermission`, `RateLimitMiddleware`).
- **Domain models:** `User`, `Session`, `Role`, `AuditLog`, `PasswordResetToken`, and `APIKey` capture the data the service manipulates. `User.Metadata` lets you attach structured context (tenant IDs, organization info, etc.) without schema changes.
- **Security helpers:** Password validation/hashing lives in `password.go`, JWT handling in `token.go`, rate limiting in `ratelimit.go`, and audit tracking in `audit.go`. Middleware and HTTP helpers wrap these components so HTTP stacks can adopt them with minimal plumbing.
- **Persistence boundaries:** All data access flows through the repository interfaces (`UserRepository`, `SessionRepository`, etc.) so you can plug in your preferred database while keeping the core logic unchanged.

## Configuration

Call `auth.LoadConfig()` at startup to gather values from the environment. Defaults are tuned for production safety but override them via env vars:

| Env var | Purpose | Default |
|---------|---------|---------|
| `AUTH_JWT_SECRET` | Signing secret for all JWT tokens | **required (no default)** |
| `AUTH_JWT_EXPIRATION` | Token lifetime (e.g., `24h`) | `24h` |
| `AUTH_JWT_ISSUER` | JWT issuer claim | `rompi-auth` |
| `AUTH_PASSWORD_MIN_LENGTH` | Minimum password length | `8` |
| `AUTH_PASSWORD_REQUIRE_UPPER` | Require uppercase chars? | `true` |
| `AUTH_PASSWORD_REQUIRE_LOWER` | Require lowercase chars? | `true` |
| `AUTH_PASSWORD_REQUIRE_NUMBER` | Require numeric chars? | `true` |
| `AUTH_PASSWORD_REQUIRE_SPECIAL` | Require symbols? | `true` |
| `AUTH_BCRYPT_COST` | bcrypt cost (4–31) | `12` |
| `AUTH_MAX_FAILED_ATTEMPTS` | How many failures before lockout | `5` |
| `AUTH_LOCKOUT_DURATION` | Lockout window (min `1m`) | `15m` |
| `AUTH_RATE_LIMIT_WINDOW` | Rate limiter window (min `1s`) | `1m` |
| `AUTH_RATE_LIMIT_MAX_REQUESTS` | Max requests per window | `5` |
| `AUTH_RESET_TOKEN_LENGTH` | Reset token length | `32` |
| `AUTH_RESET_TOKEN_EXPIRATION` | Reset token TTL | `1h` |
| `AUTH_DEFAULT_LANGUAGE` | Fallback language code | `en` |

`LoadConfig` validates every setting—missing `AUTH_JWT_SECRET`, too-short tokens, invalid durations, or a blank default language all fail fast.

## Persistence Contracts

The service delegates all storage to your implementations of:

- `UserRepository` – create/update/delete users, track failed attempts, and manage lockouts.
- `SessionRepository` – persist issued sessions so you can revoke or enumerate them.
- `RoleRepository` – manage roles, assign/remove them per user, and query permissions.
- `AuditLogRepository` – record security-relevant events for compliance and diagnostics.
- `PasswordResetTokenRepository` – create and expire reset tokens securely.
- `APIKeyRepository` – look up long-lived API keys for machine-to-machine auth.

`Repositories` bundles these interfaces for `NewService`. Only `Users` is strictly required; the rest are optional but enable features like password resets or session tracking. The `pkg/auth/testutil/mocks.go` package already implements all interfaces for tests and experimentation.

## Instantiating the Service

Wire the service once during app boot:

```go
cfg, err := auth.LoadConfig()
if err != nil {
    log.Fatalf("auth config: %v", err)
}
repos := auth.Repositories{
    Users:   userRepo,
    Sessions: sessionRepo,
    Roles:   roleRepo,
    AuditLogs: auditLogRepo,
    PasswordResetTokens: resetRepo,
    APIKeys: apiKeyRepo,
}
svc, err := auth.NewService(cfg, repos)
if err != nil { ... }
```

Handle `NewService` errors, as it verifies the config and required repositories before activating any flows.

## Routing & Middleware Integration

The service ships HTTP middleware helpers. Wrap your router as follows:

- `svc.Middleware()` validates bearer JWTs, loads the corresponding `User`, and injects it into context. Use `auth.UserFromContext(r.Context())` inside handlers to read the user.
- `svc.RequireRole("admin")` and `svc.RequirePermission("orders:write")` guard routes against insufficient privileges.
- `svc.RateLimitMiddleware()` applies the configured rate limits per origin before the handler logic runs.

Stack them once per route group and reuse the same `svc` instance; middleware is goroutine-safe.

## Authentication Flows

- **Registration:** `Register` validates email/password, hashes the password, populates `User.Language`, and stores the user. On failure it logs (via `AuditLogger`) and enforces rate limits.
- **Login:** `Login` checks credentials, enforces account lockout/failed attempts, issues a JWT via `TokenManager`, and optionally creates a session record. `LoginResponse` returns the token, expiry, and the user model.
- **Logout/Token Refresh:** `Logout` removes session records and `RefreshToken` issues a fresh JWT for a valid token.
- **Password resets:** `InitiatePasswordReset` emits a token stored via `PasswordResetTokenRepository`; `CompletePasswordReset` validates the token, enforces the password policy, updates the hash, and marks the token as used. Be sure to email the token to users securely.
- **API keys:** `ValidateAPIKey` looks up keys via `APIKeyRepository` so machine clients can authenticate without users.

### Registration example

Once your application wires up an `auth.Service`, `Register` is the entry point for onboarding new accounts:

```go
import (
	"context"
	"fmt"

	"github.com/rompi/core-backend/pkg/auth"
)

func registerUser(ctx context.Context, svc auth.Service) (*auth.User, error) {
	req := auth.RegisterRequest{
		Email:    "sam@example.com",
		Password: "Str0ngP@ssw0rd!",
		Language: "en",
	}

	user, err := svc.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}
```

Capture the returned user ID and language to seed welcome e-mails, analytics, or subsequent session issuance; the service still honors rate limits and audit logging.

All flows write concise audit events, so hooking `AuditLogRepository` to your database or log storage gives you visibility into security-sensitive actions.

## Rate Limiting & Security

`RateLimiter` is shared between login, registration, password resets, and the exposed middleware helper. You can reuse `RateLimitMiddleware` across any HTTP handler to throttle repeated abuse attempts using the same configuration.

`Password` and `Token` helpers centralize hashing and signature logic for consistent behavior across restarts. `validator.go` enforces email format and password strength based on the config.

## Localization & Errors

`i18n.go` exposes `DefaultTranslator` preloaded with English messages for every auth error code. To support other locales:

1. Register additional-language maps via `DefaultTranslator.Register(language, messages)` or load JSON files with `LoadFromFile`.
2. When returning errors use `auth.NewAuthError(code, httpStatus, language, args...)` so messages respect the caller’s locale and template placeholders (`{0}`, `{1}`) work.

Language-aware errors surface in `NewAuthError` output; `Service` defaults to `Auth.DefaultLanguage` when `Register` or password resets omit a language code.

## Testing & Examples

- Run `go test ./pkg/auth/...` to cover the service, middleware, rate limiter, password rules, and language logic. `Integration` tests exercise the SQLite helper (see `pkg/auth/integration_test.go`).
- `pkg/auth/testutil/mocks.go` gives ready-made in-memory repositories for unit tests.
- The HTTP example at `pkg/auth/examples/with-http/main.go` shows middleware stacking, handler wiring, and translator setup.

## Sample programs

Each folder under `pkg/auth/examples` demonstrates a focused scenario:

- **basic:** Shows `Register` → `Login` using the mock user repository; run `go run ./pkg/auth/examples/basic` to watch the service create a user and emit a token expiry.
- **custom-validation:** Adds bespoke password rules (e.g., containing “Rompi”) on top of the built-in validation; run `go run ./pkg/auth/examples/custom-validation` to verify both checks pass.
- **multi-language:** Registers Spanish translations, builds an error in `es`, and logs the localized text; execute `go run ./pkg/auth/examples/multi-language` to print the translated message.
- **with-http:** Wires middleware, role guards, and rate limiting into an HTTP server that protects `/protected`; start it with `go run ./pkg/auth/examples/with-http` and visit `http://localhost:8080/protected` while mocking credentials to see the middleware in action.

## Next Steps

1. Connect persistence adapters (`UserRepository`, `RoleRepository`, etc.) to your database driver.
2. Wire `svc.Middleware()` and the role/permission helpers into the routes you want to protect.
3. Emit emails/tokens via your platform when users request `InitiatePasswordReset`.
4. Feed audit logs into your observability stack to review lockouts, failed logins, and API key usage.

By following the contracts in this package, you can rely on a consistent auth surface while tailoring storage, notifications, and routing to your application.
