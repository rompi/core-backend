# Basic Example

This folder contains a minimal program that shows how to construct the auth service with the bundled test helpers and perform a straight `Register` → `Login` flow.

## Highlights

- Uses `pkg/auth/testutil.MockUserRepository` to keep all state in memory so the example is deterministic.
- Demonstrates configuring the service via `auth.Config`, validating that configuration, and calling `auth.NewService`.
- Calls `Register`, then `Login`, and prints the issued token expiry to prove the service is wired correctly.

## Running

```bash
go run ./pkg/auth/examples/basic
```

No database or external dependencies are required; the repo mocks everything needed to exercise the password validation, hashing, and JWT issuance.
