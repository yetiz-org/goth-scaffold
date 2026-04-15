# tests/e2e

End-to-end tests exercise the full HTTP stack against a running application server.
Tests must be runnable with: `make local-test-e2e`

## File Naming

```
{domain}_test.go         # e.g. health_test.go, site_settings_test.go
```

One file per logical domain. Group related endpoint tests together.

## Test Function Naming

```
Test<Domain>_<Scenario>
```

Examples:
- `TestHealth_ReturnsOK`
- `TestSiteSettings_ListReturnsEmpty`
- `TestSiteSettings_UnauthorizedRejected`

## Required Boilerplate

Every test function must:

```go
func TestFoo_Bar(t *testing.T) {
	t.Parallel()                         // MANDATORY — always first

	tc := testutils.NewTestContext(t)    // MANDATORY

	// ... test body ...
}
```

Never omit `t.Parallel()`. Never share state between test functions.

## Comments

### File Header

```go
// Tests for /api/v1/foo endpoints.
// Covers: list (GET), create (POST), not-found (404), unauthorized (401).
package e2e
```

### Test Function Comment

```go
// TestSiteSettings_ListReturnsEmptyWhenNoData verifies GET /api/v1/site-settings
// returns an empty array when no records exist, not null or an error.
func TestSiteSettings_ListReturnsEmptyWhenNoData(t *testing.T) {
```

One sentence: what endpoint, what condition, what is verified.

## Assertions

Always include a descriptive message:

```go
if resp.StatusCode != http.StatusOK {
	t.Errorf("expected 200, got %d — check if server started correctly", resp.StatusCode)
}
```

Verify response body fields, not only status codes.

## Service-Dependent Tests (Database / Redis)

Skip gracefully when the service is not configured:

```go
func TestFoo_UsesDatabase(t *testing.T) {
	t.Parallel()

	if !database.Enabled() {
		t.Skip("database not configured — skipping")
	}

	// test body using actual DB
}
```

Use `database.Enabled()` / `redis.Enabled()` from the connector packages.
Do not panic or fail on missing config — always `t.Skip`.

## Edge Cases to Cover

For every endpoint under test, cover:

| Category | Example |
|----------|---------|
| Happy path | valid input → 200 + correct body |
| Unauthorized | no/invalid token → 401 |
| Not found | nonexistent ID → 404 |
| Invalid input | empty body, malformed JSON → 400 |
| Method not allowed | wrong HTTP verb → 405 |

## Prohibited Patterns

- `time.Sleep` for async waits — use polling helpers
- Hardcoded user/resource IDs — use generated test data
- Shared state between test functions — each test is fully independent
- Assertions without messages — always explain what was expected
- Missing `t.Parallel()` — add to every test function

## Test Context Helpers

| Method | Purpose |
|--------|---------|
| `tc.DoGet(path)` | GET request to `BaseURL + path` |
| `tc.DoPostJSON(path, body)` | POST with `Content-Type: application/json` |
| `tc.NewClient()` | Fresh `http.Client` that does NOT follow redirects |
| `testutils.GetBaseURL()` | Returns current server base URL (from config) |
| `testutils.ReadJSONBody(t, resp, &dest)` | Unmarshal response body into `dest`; fails test on error |
| `testutils.GetTestConfig()` | Returns `TestConfig` with port, config path, project root |

`tc.BaseURL` is set from `testutils.GetBaseURL()` at construction — use it directly for custom requests.

## Running

```bash
# Services must be up first
make local-env-start

# Run e2e tests (automatically builds binary and starts server)
make local-test-e2e

# Or directly
SCAFFOLD_E2E_BINARY=./scaffold go test -v -count=1 -timeout=120s ./tests/e2e/...
```

`SCAFFOLD_E2E_BINARY` controls which binary the test harness starts:
- If set → use that path directly (must already be built).
- If unset → the harness builds a fresh binary from source before running tests.
- `make local-test-e2e` builds first then sets this variable automatically.
