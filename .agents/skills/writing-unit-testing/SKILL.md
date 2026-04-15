---
name: writing-unit-testing
description: Use when creating or updating unit tests for Go source files, when comprehensive unit test coverage is needed in tests/units/. Use when the user says write unit test, generate unit test, add unit test coverage, or test this file.
---

# Writing Unit Tests

## Core Rule

Every test must verify real behavior, protect a concrete invariant or regression, and pass 10 consecutive runs. No status-only assertions. No coverage-padding tests.

## When to Use

- A source file or package needs unit test coverage.
- User asks to add/update unit tests for models, helpers, services, or components.

Don't use when:
- The code requires DB, network, or app bootstrap to test — those are integration tests.
- E2E tests are needed — use `writing-e2e-testing`.

## Allowed vs Forbidden

**Allowed:**
- Pure Go logic: model helpers, validators, serialization, business rules.
- Deterministic computation with no external dependencies.
- Service logic isolated from external systems (inject mocks via `_Dependency`).

**Forbidden:**
- Real DB connections (`connector/database`, `connector/keyspaces`).
- Real network calls or Redis connections.
- `app.FlagParse()` or full app bootstrap.
- Tests that only assert `NotNil` or `NotEmpty` with no behavioral claim.
- Tests added purely for line-count or coverage metrics.

## Workflow

```
1. Read source file end-to-end
2. Classify testable units (exported funcs/methods with logic)
3. Design test matrix
4. Write test file
5. go build ./... (fix errors)
6. go test -v -count=1 ./tests/units/... (run 10x)
```

### 1. Read the Source File

For each exported function/method extract:
- **Signature**: inputs, outputs, receiver
- **Logic branches**: every `if`, `switch`, `for`
- **Edge cases**: nil receiver, zero values, empty strings, boundary values
- **Dependencies**: does it call external systems? If yes → integration test, not unit test

### 2. Classify Testable Units

| Testable | Not Testable (unit) |
|----------|---------------------|
| Pure helpers, validators | DB queries |
| Model methods | Network calls |
| Serialization/deserialization | File I/O in prod paths |
| Business rule branches | Service with uninjectable deps |

### 3. Design Test Matrix

For each function, cover ALL of:

| Category | Examples |
|----------|---------|
| Happy path | valid input → expected output |
| Nil/zero receiver | should not panic |
| Empty input | empty string, nil slice |
| Boundary values | max length, edge IDs |
| Error paths | invalid input, wrong type |

### 4. Write the Test File

```go
// Whitebox test — same package grants access to unexported fields (_Dependency, _FooDeps).
// Use `package services_test` for blackbox tests that only call exported methods.
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFooService_ListItems_ReturnsEmpty(t *testing.T) {
    svc := &FooService{
        _Dependency: &_FooDeps{
            FooRepo: &mockFooRepo{items: nil},
        },
    }

    result := svc.ListItems()

    assert.NotNil(t, result, "ListItems should return empty slice, not nil")
    assert.Empty(t, result)
}
```

Rules:
- One `Test` function per behavior, not per method.
- Use table-driven tests for multiple input/output combinations.
- Test function name: `Test<Type>_<Method>_<Scenario>`.
- Never share state between test functions.

### 5. Verify

```bash
go build ./...
go test -v -count=1 ./tests/units/...
```

Run 10 consecutive times — if any run fails, fix root cause and restart count.

## Test File Placement

```
tests/units/
├── models/          # model method tests
├── services/        # service logic tests (with mocked deps)
├── components/      # helper/component tests
└── conf/            # config parsing tests
```

Keep tests under `tests/units/`, not co-located with source files.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Testing only `NotNil` | Assert specific expected values |
| Importing real DB connectors | Inject mocks via `_Dependency` pattern |
| Testing private implementation details | Test through exported interface |
| `t.Log` without assertion | Assertions only — logs don't fail tests |
| Sharing state between tests | Each test creates its own instance |
| Missing nil receiver test | Add explicit nil receiver case for pointer methods |
