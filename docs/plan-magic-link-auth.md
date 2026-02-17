# Magic Link Authentication — Implementation Plan

## Overview

Add passwordless magic link login to the existing `pkg/auth` package. A user provides their email address, receives a one-time link containing a cryptographically random token, and clicking that link authenticates them — no password required.

The design follows the same patterns already established in the codebase: interface-based repositories, environment-driven configuration, rate limiting, audit logging, and transport-agnostic service methods.

---

## 1. Scope

### In Scope

- New `MagicLinkToken` model and `MagicLinkTokenRepository` interface
- New configuration fields for magic link token length, expiration, and feature toggle
- Two new `Service` methods: `InitiateMagicLink` and `CompleteMagicLink`
- Rate limiting on magic link requests (per email)
- Audit logging for initiation and completion events
- New error codes and i18n messages for magic link flows
- Mock repository for unit testing
- Unit tests for the full magic link lifecycle
- Integration with existing JWT/session issuance on successful verification

### Out of Scope

- Email delivery (the service returns the token; the caller is responsible for sending the email — consistent with how `InitiatePasswordReset` works today)
- Frontend / UI changes
- Database migration scripts (repository is interface-based; implementors create their own schema)

---

## 2. Affected Files

| File | Change Type | Description |
|---|---|---|
| `pkg/auth/models.go` | **Modify** | Add `MagicLinkToken` struct |
| `pkg/auth/repository.go` | **Modify** | Add `MagicLinkTokenRepository` interface |
| `pkg/auth/service.go` | **Modify** | Add `InitiateMagicLink` and `CompleteMagicLink` to `Service` interface; add `MagicLinkRequest` type |
| `pkg/auth/service_impl.go` | **Modify** | Add `MagicLinkTokens` to `Repositories` struct; implement `InitiateMagicLink` and `CompleteMagicLink` |
| `pkg/auth/config.go` | **Modify** | Add `MagicLinkTokenLength`, `MagicLinkTokenExpiration`, `MagicLinkEnabled` fields; env loading and validation |
| `pkg/auth/errors.go` | **Modify** | Add `CodeInvalidMagicLink`, `CodeMagicLinkDisabled`, and corresponding sentinel errors |
| `pkg/auth/i18n.go` | **Modify** | Add English translations for new error codes |
| `pkg/auth/testutil/mocks.go` | **Modify** | Add `MockMagicLinkTokenRepository` |
| `pkg/auth/magic_link_test.go` | **Create** | Unit tests for magic link initiation, verification, expiry, reuse prevention, and rate limiting |

---

## 3. Detailed Design

### 3.1 Model — `MagicLinkToken`

Add to `pkg/auth/models.go`:

```go
// MagicLinkToken represents a one-time token sent to a user's email
// for passwordless authentication.
type MagicLinkToken struct {
    Token     string    `json:"token"`
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    IssuedAt  time.Time `json:"issued_at"`
    ExpiresAt time.Time `json:"expires_at"`
    Used      bool      `json:"used"`
}
```

**Design notes:**
- `Email` is stored on the token itself so verification can proceed without a separate user lookup when the user does not yet exist (supports optional auto-registration — see Section 6).
- `Used` flag prevents token reuse after successful authentication.
- Follows the same pattern as the existing `PasswordResetToken`.

### 3.2 Repository — `MagicLinkTokenRepository`

Add to `pkg/auth/repository.go`:

```go
// MagicLinkTokenRepository defines persistence for magic link tokens.
type MagicLinkTokenRepository interface {
    Create(ctx context.Context, token *MagicLinkToken) error
    GetByToken(ctx context.Context, token string) (*MagicLinkToken, error)
    MarkUsed(ctx context.Context, token string) error
    DeleteByUserID(ctx context.Context, userID string) error
    DeleteExpired(ctx context.Context) error
}
```

**Method rationale:**
- `MarkUsed` — atomically marks a token as consumed (avoids race conditions vs. a read-modify-write with `Update`).
- `DeleteByUserID` — invalidates any outstanding tokens for a user when a new one is issued (only one active magic link per user at a time).
- `DeleteExpired` — cleanup hook, same pattern as `PasswordResetTokenRepository` and `SessionRepository`.

### 3.3 Configuration

Add to `pkg/auth/config.go`:

```go
// New Config fields
MagicLinkEnabled         bool          `json:"magic_link_enabled"`
MagicLinkTokenLength     int           `json:"magic_link_token_length"`
MagicLinkTokenExpiration time.Duration `json:"magic_link_token_expiration"`
```

| Environment Variable | Default | Validation |
|---|---|---|
| `AUTH_MAGIC_LINK_ENABLED` | `false` | Boolean |
| `AUTH_MAGIC_LINK_TOKEN_LENGTH` | `32` | Minimum 16 |
| `AUTH_MAGIC_LINK_TOKEN_EXPIRATION` | `15m` | Minimum 1m, maximum 1h |

**Defaults rationale:**
- 15-minute expiration balances usability (time to check email) with security (short window of exposure).
- 32-character hex token provides 128 bits of entropy — sufficient for a one-time-use credential.
- Disabled by default so existing deployments are unaffected until explicitly opted in.

### 3.4 Error Codes

Add to `pkg/auth/errors.go`:

```go
const (
    CodeInvalidMagicLink  = "invalid_magic_link"
    CodeMagicLinkDisabled = "magic_link_disabled"
)

var (
    ErrInvalidMagicLink  = errors.New("invalid or expired magic link token")
    ErrMagicLinkDisabled = errors.New("magic link authentication is not enabled")
)
```

### 3.5 Service Interface Changes

Add to `pkg/auth/service.go`:

```go
// New request type
type MagicLinkRequest struct {
    Email    string `json:"email"`
    Language string `json:"language"`
}

// New methods on Service interface
InitiateMagicLink(ctx context.Context, req MagicLinkRequest) (*MagicLinkToken, error)
CompleteMagicLink(ctx context.Context, token string) (*LoginResponse, error)
```

### 3.6 Service Implementation

Add to `pkg/auth/service_impl.go`:

#### `InitiateMagicLink`

```
1. Validate that magic link is enabled in config (return ErrMagicLinkDisabled if not)
2. Validate the MagicLinkTokenRepository is configured (return error if nil)
3. Normalize and validate email
4. Rate limit on key "magic_link:{email}"
5. Look up user by email
   - If user not found → return ErrUserNotFound
     (auto-registration is a future extension — see Section 6)
6. Delete any existing magic link tokens for this user (DeleteByUserID)
7. Generate cryptographically random token via existing generateRandomToken()
8. Create MagicLinkToken with expiration from config
9. Persist token via repository
10. Audit log: action "magic_link_initiated"
11. Return *MagicLinkToken (caller sends email with the token)
```

#### `CompleteMagicLink`

```
1. Validate token is non-empty
2. Validate that magic link is enabled in config
3. Validate the MagicLinkTokenRepository is configured
4. Fetch token from repository
5. Validate: token exists, not used, not expired
   - If invalid → return ErrInvalidMagicLink
6. Mark token as used (MarkUsed)
7. Fetch user by token.UserID
8. Check account is not locked (LockedUntil)
   - If locked → return ErrAccountLocked
9. Reset failed login attempts (successful magic link clears lockout)
10. Generate JWT via tokenManager.Generate(user)
11. Create session if SessionRepository is configured
12. Audit log: action "magic_link_completed"
13. Return *LoginResponse (same shape as password-based Login)
```

### 3.7 Repositories Struct Update

Update `Repositories` in `pkg/auth/service_impl.go`:

```go
type Repositories struct {
    Users               UserRepository
    Sessions            SessionRepository
    Roles               RoleRepository
    AuditLogs           AuditLogRepository
    PasswordResetTokens PasswordResetTokenRepository
    APIKeys             APIKeyRepository
    MagicLinkTokens     MagicLinkTokenRepository      // new
}
```

The `validate()` method does **not** require `MagicLinkTokens` — it is optional, same as `Sessions`, `Roles`, etc. The `InitiateMagicLink` and `CompleteMagicLink` methods check for `nil` at runtime and return a descriptive error.

### 3.8 i18n Messages

Add English translations in `pkg/auth/i18n.go`:

```go
"invalid_magic_link":  "The magic link is invalid or has expired.",
"magic_link_disabled": "Magic link authentication is not enabled.",
```

### 3.9 Mock Repository

Add to `pkg/auth/testutil/mocks.go`:

```go
type MockMagicLinkTokenRepository struct {
    CreateFunc        func(ctx context.Context, token *auth.MagicLinkToken) error
    GetByTokenFunc    func(ctx context.Context, token string) (*auth.MagicLinkToken, error)
    MarkUsedFunc      func(ctx context.Context, token string) error
    DeleteByUserIDFunc func(ctx context.Context, userID string) error
    DeleteExpiredFunc func(ctx context.Context) error
}
```

Follows the same function-pointer pattern used by all other mock repositories in the testutil package.

---

## 4. Test Plan

### Unit Tests (`pkg/auth/magic_link_test.go`)

| Test Case | Description |
|---|---|
| `TestInitiateMagicLink_Success` | Valid email for existing user → returns token, audit logged |
| `TestInitiateMagicLink_DisabledConfig` | Magic link disabled → returns `ErrMagicLinkDisabled` |
| `TestInitiateMagicLink_NilRepository` | Repository not configured → returns descriptive error |
| `TestInitiateMagicLink_InvalidEmail` | Malformed email → returns validation error |
| `TestInitiateMagicLink_UserNotFound` | Unknown email → returns `ErrUserNotFound` |
| `TestInitiateMagicLink_RateLimited` | Exceeds rate limit → returns `ErrRateLimitExceeded` |
| `TestInitiateMagicLink_InvalidatesOldTokens` | Calling twice deletes the first token |
| `TestCompleteMagicLink_Success` | Valid token → returns JWT + session, token marked used |
| `TestCompleteMagicLink_ExpiredToken` | Expired token → returns `ErrInvalidMagicLink` |
| `TestCompleteMagicLink_UsedToken` | Already-used token → returns `ErrInvalidMagicLink` |
| `TestCompleteMagicLink_NonexistentToken` | Unknown token → returns `ErrInvalidMagicLink` |
| `TestCompleteMagicLink_LockedAccount` | User locked → returns `ErrAccountLocked` |
| `TestCompleteMagicLink_ResetsFailedAttempts` | Successful magic link clears lockout state |
| `TestCompleteMagicLink_CreatesSession` | Session created when SessionRepository is configured |
| `TestCompleteMagicLink_NoSession` | Works without SessionRepository (nil) |

---

## 5. Security Considerations

| Concern | Mitigation |
|---|---|
| **Token brute-force** | 32-char hex = 128 bits of entropy; rate limiting on `magic_link:{email}` |
| **Token replay** | One-time use: `Used` flag set atomically via `MarkUsed` |
| **Token leakage via URL** | Tokens should be placed in URL fragment or POST body by the frontend; this is a caller responsibility but should be documented |
| **Enumeration attack** | `InitiateMagicLink` returns `ErrUserNotFound` for unknown emails. Consider returning a generic success response at the HTTP handler layer to avoid revealing whether an email is registered (handler-level concern, not service-level) |
| **Concurrent requests** | `DeleteByUserID` before creating a new token ensures only one valid token exists per user at any time |
| **Short expiration** | Default 15 minutes; maximum configurable value capped at 1 hour |
| **Account lockout bypass** | `CompleteMagicLink` checks `LockedUntil` before issuing a JWT — locked accounts cannot authenticate via magic link |

---

## 6. Future Extensions (Out of Scope)

These are not part of this implementation but the design accommodates them:

- **Auto-registration**: If the user does not exist, `InitiateMagicLink` could optionally create a new account with a nil password hash. The `MagicLinkToken.Email` field supports this without requiring a `UserID` at creation time.
- **Email delivery abstraction**: An `EmailSender` interface could be introduced to let the auth package send emails directly, removing the burden from callers.
- **Magic link for account verification**: Newly registered users could receive a magic link to verify their email address, combining signup confirmation and first login.
- **Device fingerprinting**: Bind the magic link token to a device fingerprint to prevent cross-device token theft.

---

## 7. Implementation Order

The recommended implementation sequence, with each step building on the previous:

1. **Config** — Add new fields to `Config`, environment variable loading, and validation in `config.go`
2. **Model** — Add `MagicLinkToken` struct to `models.go`
3. **Errors** — Add error codes and sentinel errors to `errors.go`; add i18n messages to `i18n.go`
4. **Repository** — Add `MagicLinkTokenRepository` interface to `repository.go`
5. **Mock** — Add `MockMagicLinkTokenRepository` to `testutil/mocks.go`
6. **Service interface** — Add `MagicLinkRequest`, `InitiateMagicLink`, and `CompleteMagicLink` to `service.go`
7. **Service implementation** — Add `MagicLinkTokens` to `Repositories`; implement both methods in `service_impl.go`
8. **Tests** — Write unit tests in `magic_link_test.go` covering all cases from Section 4

---

## 8. Reference: Database Schema (for implementors)

Implementors of `MagicLinkTokenRepository` will need a table similar to:

```sql
CREATE TABLE magic_link_tokens (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    email      TEXT NOT NULL,
    issued_at  INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_magic_link_tokens_user_id ON magic_link_tokens(user_id);
CREATE INDEX idx_magic_link_tokens_expires_at ON magic_link_tokens(expires_at);
```
