# Multi-language Example

This example shows how to customize the translator that the auth package uses to produce localized error messages.

## Highlights

- Creates a new `auth.Translator`, registers Spanish translations for a few error codes, and then swaps the package-level `auth.DefaultTranslator`.
- Calls `auth.NewAuthError` with the Spanish locale to show how error messages surface in another language.
- Keeps the example self-contained (no repositories or HTTP handlers) so you can see the string override logic clearly.

## Running

```bash
go run ./pkg/auth/examples/multi-language
```

Inspect the log output to verify the Spanish message and adapt the translator registration for any additional locales you need.
