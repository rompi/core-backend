# Auth Package API

## Service

`github.com/rompi/core-backend/pkg/auth.Service` exposes the following methods:

- `Register(ctx, RegisterRequest)` – creates a new user (email/password) with password complexity checks.
- `Login(ctx, LoginRequest)` – authenticates a user, stores a session (if repository provided), and returns a JWT/expiration.
- `Logout(ctx, token)` – clears the session tied to `token`.
- `ValidateToken(ctx, token)` / `RefreshToken(ctx, token)` – decode/renew JWTs via `TokenManager`.
- `InitiatePasswordReset(ctx, email)` / `CompletePasswordReset(ctx, token, newPassword)` – issue tokens and allow password updates.
- `ChangePassword(ctx, userID, oldPassword, newPassword)` – update an existing account password.
- `ValidateAPIKey(ctx, apiKey)` – resolve a stored API key to its user and ensure it has not expired or been revoked.
- `GetUserRoles(ctx, userID)` / `CheckPermission(ctx, userID, permission)` – inspect user roles and permissions.
- Middleware helpers (`Middleware`, `RequireRole`, `RequirePermission`, `RateLimitMiddleware`) for HTTP servers.

## Models

- `Config` (see `pkg/auth/config.go`) determines JWT secrets, password rules, lockout thresholds, and rate-limiting windows.
- `User`, `Session`, `Role`, `PasswordResetToken`, `APIKey` models mirror the fields stored by consumer repositories.
- `AuthError` enumerates known error codes (`CodeInvalidCredentials`, `CodeUserNotFound`, etc.) with translation support.

## Error Handling

Use `NewAuthError(code, status, language, details...)` whenever an API handler needs to return a structured error response. The translator hierarchy (English default + optional `Register`/`LoadFromFile`) ensures a localized message is always available.

## Integration Guidance

- Repository interfaces (`repository.go`) are intentionally minimal so consumers plug in their favorite persistence and still reuse the auth logic.
- The included `testutil` mocks demonstrate how to build unit tests without hitting a real database.
- See `docs/INTEGRATION.md` for instructions on running the SQLite-based integration test harness.
