# Plan: Fix Session ID Collision (Issue #15)

## Issue Summary

**GitHub Issue**: [#15 - Session ID collision causes login failures after password change](https://github.com/rompi/core-backend/issues/15)

**Symptom**: Users experience login failures with error `"ERROR: duplicate key value violates unique constraint 'sessions_pkey'"` (SQLSTATE 23505) when changing password and immediately logging in.

---

## Root Cause Analysis

### Current Implementation

JWT tokens are generated in `pkg/auth/token.go:35-54`:

```go
func (m *TokenManager) Generate(user *User) (string, time.Time, error) {
    now := time.Now().UTC()
    expiration := now.Add(m.expiration)
    claims := Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    m.issuer,
            Subject:   user.ID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(expiration),
        },
        UserID: user.ID,
        Email:  user.Email,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(m.secret)
    // ...
}
```

### The Problem

JWT tokens are **deterministic** - they are derived entirely from:

| Input | Value |
|-------|-------|
| Issuer | Constant (from config) |
| Subject | User ID (same for same user) |
| IssuedAt | Unix timestamp (second precision) |
| ExpiresAt | Derived from IssuedAt |
| UserID | Same as Subject |
| Email | User's email (constant) |
| Signing Secret | Constant |

**When two tokens are generated for the same user within the same second, they are identical.**

### Collision Scenario

```
Timeline (within 1 second):
├─ 14:30:05.100 - Password change completes
├─ 14:30:05.300 - Login request arrives
├─ 14:30:05.400 - Token generated (IssuedAt = 1706972405)
├─ 14:30:05.500 - Session created with token as PRIMARY KEY
├─ 14:30:05.600 - Second login (or refresh) request
├─ 14:30:05.700 - Token generated (IssuedAt = 1706972405) ← IDENTICAL
└─ 14:30:05.800 - Session insert fails: duplicate key violation
```

### Database Schema

From `pkg/auth/integration_test.go:95-102`:

```sql
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,  -- JWT token used as primary key
    user_id TEXT,
    issued_at INTEGER,
    expires_at INTEGER,
    revoked INTEGER,
    metadata TEXT
)
```

The session table uses the JWT token as the primary key, so duplicate tokens cause constraint violations.

---

## Proposed Solution

**Add a `jti` (JWT ID) claim with a UUID** to ensure cryptographic uniqueness.

### Why This Approach?

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **UUID jti claim** | Standard JWT practice (RFC 7519), cryptographically unique, minimal change | Slightly larger token (~36 chars) | ✅ **Selected** |
| Nanosecond timestamp | Simple change | Still has collision risk under high concurrency | ❌ Rejected |
| Retry on constraint failure | Works around the issue | Adds complexity, masks root cause, poor UX | ❌ Rejected |
| Separate session ID column | Decouples session from token | Requires schema migration, larger change | ❌ Rejected |

---

## Implementation Plan

### Step 1: Update Token Generation

**File**: `pkg/auth/token.go`

Add UUID import and `jti` claim:

```go
import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"  // ADD THIS
)

func (m *TokenManager) Generate(user *User) (string, time.Time, error) {
    now := time.Now().UTC()
    expiration := now.Add(m.expiration)
    claims := Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        uuid.NewString(),  // ADD THIS - unique JWT ID (jti claim)
            Issuer:    m.issuer,
            Subject:   user.ID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(expiration),
        },
        UserID: user.ID,
        Email:  user.Email,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(m.secret)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("signing token: %w", err)
    }
    return signed, expiration, nil
}
```

### Step 2: Add Unit Test for Token Uniqueness

**File**: `pkg/auth/token_test.go`

```go
func TestTokenManager_Generate_UniqueTokens(t *testing.T) {
    cfg := &Config{
        JWTSecret:             "test-secret",
        JWTIssuer:             "test-issuer",
        JWTExpirationDuration: time.Hour,
    }
    tm := NewTokenManager(cfg)
    user := &User{ID: "user-123", Email: "test@example.com"}

    // Generate multiple tokens for the same user in rapid succession
    tokens := make(map[string]bool)
    for i := 0; i < 100; i++ {
        token, _, err := tm.Generate(user)
        if err != nil {
            t.Fatalf("failed to generate token: %v", err)
        }
        if tokens[token] {
            t.Fatalf("duplicate token generated on iteration %d", i)
        }
        tokens[token] = true
    }
}
```

### Step 3: Add Integration Test for Rapid Login

**File**: `pkg/auth/service_impl_test.go`

```go
func TestService_RapidLoginAfterPasswordChange(t *testing.T) {
    // Setup service with test repositories
    svc, repos := setupTestService(t)
    ctx := context.Background()

    // Register user
    user, err := svc.Register(ctx, RegisterRequest{
        Email:    "test@example.com",
        Password: "OldPassword123!",
    })
    require.NoError(t, err)

    // Change password
    err = svc.ChangePassword(ctx, user.ID, "OldPassword123!", "NewPassword456!")
    require.NoError(t, err)

    // Immediately login multiple times (simulating rapid requests)
    for i := 0; i < 5; i++ {
        resp, err := svc.Login(ctx, LoginRequest{
            Email:    "test@example.com",
            Password: "NewPassword456!",
        })
        require.NoError(t, err, "login attempt %d failed", i+1)
        require.NotEmpty(t, resp.Token)
    }

    // Verify all sessions are unique
    sessions, err := repos.Sessions.GetByUserID(ctx, user.ID)
    require.NoError(t, err)

    tokenSet := make(map[string]bool)
    for _, s := range sessions {
        require.False(t, tokenSet[s.Token], "duplicate session token found")
        tokenSet[s.Token] = true
    }
}
```

---

## Files to Modify

| File | Change | Lines Affected |
|------|--------|----------------|
| `pkg/auth/token.go` | Add uuid import, add `ID: uuid.NewString()` to claims | ~3 lines |
| `pkg/auth/token_test.go` | Add uniqueness test | ~20 lines |
| `pkg/auth/service_impl_test.go` | Add rapid login integration test | ~35 lines |

---

## Verification

### Manual Testing

1. Register a new user
2. Change password
3. Immediately login (within 1 second)
4. Verify login succeeds without 500 error

### Automated Testing

```bash
go test ./pkg/auth/... -v -run "TestTokenManager_Generate_UniqueTokens|TestService_RapidLoginAfterPasswordChange"
```

---

## Rollback Plan

If issues arise, revert the single commit that adds the `jti` claim. Existing tokens remain valid as the change only affects token generation, not validation.

---

## References

- [RFC 7519 - JSON Web Token (JWT)](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7) - `jti` claim specification
- [google/uuid](https://github.com/google/uuid) - UUID library already used in codebase
