# HTTP Middleware Example

This example wires the auth service directly into an HTTP server to demonstrate middleware, role checking, and rate limiting.

## Highlights

- Configures `auth.Config` and registers `testutil.MockUserRepository` + `testutil.MockRoleRepository` so the service can resolve a demo user and role.
- Protects `/protected` with `svc.Middleware()` and `svc.RequireRole("admin")`, then uses `svc.RateLimitMiddleware()` on the whole mux.
- Shows how to read the authenticated user inside handlers via `auth.UserFromContext`.
- Runs a simple `http.Server` on `:8080`, printing a greeting once authentication succeeds.

## Running

```bash
go run ./pkg/auth/examples/with-http
```

After starting, call `curl -H "Authorization: Bearer <token>" http://localhost:8080/protected` (use the token emitted by `auth.NewService` from your own client) to see the handler respond.
