# Integration & Benchmark Guidance

## Integration Tests

The package ships with an SQLite-backed integration test (`pkg/auth/integration_test.go`). It exercises the auth service fully (register/login/password reset/change) against a real database schema. Run it with:

```bash
go test ./pkg/auth -run Integration
```

The test uses `modernc.org/sqlite` so it can run in pure Go without cgo.

## Benchmarks

`pkg/auth/bench_test.go` contains benchmark helpers for password hashing and JWT generation to gauge the performance impact of configuration tweaks. Run the suite combined with profiling options as needed:

```bash
go test ./pkg/auth -bench=. -benchmem
```
