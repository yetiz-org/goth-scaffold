# tests

## Mandatory Flags

All test runs must use `-v -count=1`:

```bash
go test -v -count=1 ./tests/units/...          # unit tests (no services required)
go test -v -count=1 -race ./tests/units/...    # with race detector (see below)
go test -v -count=1 -timeout=120s ./tests/e2e/... # e2e (services + binary required)
```

| Flag | Effect | When required |
|------|--------|---------------|
| `-v` | Show individual test names and PASS/FAIL lines | Always |
| `-count=1` | Disable Go's test result cache — always re-executes | Always |
| `-race` | Enable the race detector — catches concurrent map/variable access | All unit tests; skip for e2e (too slow) |
| `-timeout=120s` | Hard deadline per test binary | E2E only |

**Never pipe test output through `grep`, `tail`, or `/dev/null`.**
Full output must be visible for diagnosis.

**Why `-count=1` matters**: Go caches passing test results by default. Without it, a passing result may hide a regression introduced since the last run.

## Directory Layout

```
tests/
├── units/          # pure Go, no external services required
│   ├── conf/
│   ├── models/
│   ├── services/
│   ├── repositories/
│   └── helpers/
└── e2e/            # full HTTP stack tests (requires services + config.yaml.local)
    └── testutils/
```

## Test File Naming

- Unit: `_test.go` under `tests/units/<package>/`
- E2E: `tests/e2e/`

## Coverage

```bash
go test -v -count=1 -coverprofile=coverage.out ./tests/units/...
go tool cover -html=coverage.out -o coverage.html
```

Minimum 80% coverage for non-example code. There is no `make test-cover` target — run the commands above directly.
