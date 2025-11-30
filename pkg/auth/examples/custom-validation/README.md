# Custom Validation Example

Use this example to learn how to layer your own password requirements on top of the defaults provided by `auth.ValidatePassword`.

## Highlights

- Builds a bespoke `auth.Config` that requires extra length but disables the built-in special-character rule so the example can enforce it manually.
- Calls `auth.ValidatePassword` with the config, then performs an additional check (`strings.Contains(password, "Rompi")`) before proceeding.
- Instantiates `auth.NewService` with a no-op `UserRepository` implementation to keep the example focused on validation logic.

## Running

```bash
go run ./pkg/auth/examples/custom-validation
```

This demonstrates how to mix the package’s policy helpers with your own business rules before accepting a password.
