# Auth Package Examples

- `pkg/auth/examples/basic/` – console program showing registration and login with the auth service and mock repositories.
- `pkg/auth/examples/custom-validation/` – illustrates how to combine built-in password validation with an additional custom rule before calling `Register`.
- `pkg/auth/examples/multi-language/` – configures the translator, loads Spanish messages, and renders a localized `AuthError`.
- `pkg/auth/examples/with-http/` – an HTTP server wiring the middleware helpers (`Middleware`, `RequireRole`, `RateLimitMiddleware`).

Run any example with `go run ./pkg/auth/examples/<example>` once the module dependencies are available.
