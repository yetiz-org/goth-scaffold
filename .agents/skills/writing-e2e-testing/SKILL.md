---
name: writing-e2e-testing
description: Use when creating or updating E2E tests for HTTP handlers in tests/e2e/. Use when the user says write e2e, generate e2e test, add e2e coverage, or test this handler.
---

# Writing E2E Tests

## Core Rule

Every test verifies real business logic through the full HTTP stack — not just status codes.
Tests must be parallel-safe, fully independent, and pass 10 consecutive runs.

## When to Use

- A handler file or endpoint path needs E2E coverage.
- User asks to add E2E tests for a specific route.

Don't use when:
- Unit tests only — use `writing-unit-testing`.
- The handler doesn't exist yet — write the handler first.

## Workflow

```
1. Identify handler and HTTP methods
2. Trace route chain (route.go → acceptances → handler)
3. Read every implemented method
4. Design test matrix
5. Write test file
6. go build ./tests/e2e/...
7. Run tests 10 consecutive times
```

### Phase 1: Handler Analysis

#### 1. Identify Implemented Methods

Read the handler file. Record which HTTP methods are actually implemented (not just inherited defaults):
- `Index` (GET list), `Get` (GET single), `Post`, `Patch`, `Put`, `Delete`
- Skip methods that return `MethodNotAllowed` or `NotImplemented` (inherited defaults).

#### 2. Trace Route Chain

Open `app/handlers/route.go`. Find the `SetEndpoint(...)` for this handler.
Record the full acceptance chain from root to endpoint — this tells you auth requirements and pre-populated params.

#### 3. Read Each Implemented Method

Extract:
- **Request validation**: body check, JSON unmarshal, required fields
- **Permission checks**: beyond route-level
- **Business logic**: service/repository calls, side effects
- **Response shape**: exact JSON structure
- **Error paths**: every error return and the condition that triggers it

### Phase 2: Test Design

For each implemented method, cover ALL categories:

| Category | Mandatory |
|----------|-----------|
| Happy path — verify full response body | YES |
| Unauthorized — no/invalid token | YES (if auth-gated) |
| Forbidden — wrong permission level | YES (if permission-gated) |
| Input validation — empty body, malformed JSON | YES (if accepts body) |
| Not found — nonexistent resource ID | YES (if path has IDs) |
| Method not allowed — wrong HTTP verb | YES |

### Phase 3: Write Test File

```go
// Tests for GET /api/v1/foos and POST /api/v1/foos.
// Covers: list, create, unauthorized, invalid-input.
package e2e

import (
    "net/http"
    "testing"

    "github.com/example/myapp/tests/e2e/testutils"
)

func TestFoos_List_ReturnsOK(t *testing.T) {
    t.Parallel()

    tc := testutils.NewTestContext(t)

    resp, err := tc.DoGet("/api/v1/foos")
    if err != nil {
        t.Fatalf("GET /api/v1/foos failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }

    // Parse and verify body fields
    var body struct {
        Items []struct{ Name string `json:"name"` } `json:"items"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
}

func TestFoos_List_Unauthorized(t *testing.T) {
    t.Parallel()
    tc := testutils.NewTestContext(t)

    resp, err := tc.DoGet("/api/v1/foos")
    if err != nil {
        t.Fatalf("GET /api/v1/foos failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", resp.StatusCode)
    }
}
```

Rules:
- `t.Parallel()` — mandatory, always first.
- `testutils.NewTestContext(t)` — mandatory.
- Always `defer resp.Body.Close()`.
- Every assertion includes a descriptive message.
- Verify response body fields, not only status codes.

## Service-Dependent Tests

```go
func TestFoos_DatabaseWrite(t *testing.T) {
    t.Parallel()

    if !database.Enabled() {
        t.Skip("database not configured — skipping")
    }

    tc := testutils.NewTestContext(t)
    // ... test body ...
}
```

Use `database.Enabled()` / `redis.Enabled()` to skip gracefully when services are absent.

## Test Helpers

| Method | Purpose |
|--------|---------|
| `tc.DoGet(path)` | GET request |
| `tc.DoPostJSON(path, body)` | POST with JSON body |
| `tc.NewClient()` | Fresh client (no redirect follow) |
| `testutils.GetBaseURL()` | Current server base URL |

## Run & Verify

```bash
go build ./tests/e2e/...    # compile check

# Run 10 consecutive times
for i in $(seq 1 10); do
  echo "=== Run $i/10 ==="
  go test -v -count=1 -timeout=120s ./tests/e2e/... -run "TestFoos_"
  [ $? -ne 0 ] && echo "FAILED on run $i" && break
done
```

If any run fails — fix root cause and restart count from 1.

## Prohibited Patterns

| Pattern | Fix |
|---------|-----|
| Status-code-only assertions | Parse and verify response body |
| `>= 200 && < 500` range checks | Assert exact expected status |
| Hardcoded resource IDs | Generate unique test data per run |
| Missing `t.Parallel()` | Add to every test function |
| `time.Sleep` for async | Use polling helpers |
| Shared state between tests | Each test is fully independent |
| Assertions without messages | Always include descriptive message |

## Checklist

- [ ] `t.Parallel()` in every test
- [ ] Happy path verifies response body, not just status
- [ ] Unauthorized test present (if auth-gated)
- [ ] Input validation test present (if accepts body)
- [ ] `go build ./tests/e2e/...` passes
- [ ] 10 consecutive test runs pass
