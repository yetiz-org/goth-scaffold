---
description: E2E Test Generation - Complete AI Agent Workflow
auto_execution_mode: 3
---

# E2E Testing Workflow

## GOLDEN RULES
1. **Verify before use**: Check `tests/e2e/testutils/*.go` for function existence
2. **Copy existing patterns**: Reference actual tests in `tests/e2e/*_test.go`
3. **Run tests**: Work incomplete until `go test -v ./...` shows 100% PASS
4. **No fabrication**: Use only functions listed in API Reference (Section B)
5. **All content in English**: All comments, logs, and documentation must be in English

---

## WORKFLOW

### 1. Classify Test Type
- **E2E**: Multi-step flow, session persistence, cross-database operations
- **Unit Test**: Pure logic
- **Integration Test**: Single DB query

### 2. Select Dependencies
| Need | Use |
|------|-----|
| Registered user | `CompleteSignup(t, baseURL)` |
| API token | OAuth flow (Section D) |
| Custom flow | `Do*` helpers (Section B) |
| No user | `NewTestClient()` |

### 3. Generate Code
// turbo
1. Create `tests/e2e/{feature}_test.go`
2. Use template from Section A
3. Name: `Test<Feature>_<Scenario>`

### 4. Validate Checklist
- [ ] `GenerateTestEmail()` - no hardcoded emails
- [ ] `baseURL + "/path"` - no hardcoded URLs
- [ ] `defer user.Cleanup(t)` for CompleteSignup
- [ ] `defer resp.Body.Close()` on all HTTP calls
- [ ] `require.NoError(t, err)` for errors
- [ ] Success + error test cases

### 5. Run & Fix
// turbo
```bash
go test -v ./...
```
Fix failures → check Section C → re-run until 100% PASS

---

## REFERENCE

### A. Template

```go
// Package e2e - {Feature} Test Suite
package e2e

// Test{Feature}_Success tests successful {feature} execution
func Test{Feature}_Success(t *testing.T) {
    user := testutils.CompleteSignup(t, baseURL)
    defer user.Cleanup(t)

    form := url.Values{"key": {"value"}}
    resp, err := user.Client.PostForm(baseURL+"/endpoint", form)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test{Feature}_Error tests error handling
func Test{Feature}_Error(t *testing.T) {
    client := testutils.NewTestClient()

    resp, err := client.PostForm(baseURL+"/endpoint", url.Values{})
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
```

### B. testutils API

| Function | Returns |
|----------|---------|
| `CompleteSignup(t, baseURL)` | `*SignupResult` (has `.Client`, `.Email`, `.Cleanup(t)`) |
| `NewTestClient()` | `*http.Client` with cookiejar |
| `GenerateTestEmail()` | `"test_{uuid}@test.local"` |
| `GenerateTestData()` | `*TestData` |
| `CleanupUser(t, email)` | Removes user from DB |
| `GetServerLogOffset()` | `int64` |
| `GetVerificationCodeFromServerLogAfterOffset(t, email, timeout, offset)` | `string` (6 digits) |
| `DoSignupBegin(t, client, baseURL, email, password)` | `*http.Response` |
| `DoSignupVerify(t, client, baseURL, code)` | `*http.Response` |
| `DoSignupComplete(t, client, baseURL)` | `*http.Response` |
| `ReadResponseBody(t, resp)` | `string` |
| `ExtractCSRFToken(html)` | `string` |
| `GetSessionID(client, baseURL)` | `string` |

### C. Common Fixes

| ❌ Wrong | ✅ Correct |
|---------|-----------|
| `email := "test@test.com"` | `email := testutils.GenerateTestEmail()` |
| `client.Get("http://localhost:8080/x")` | `user.Client.Get(baseURL + "/x")` |
| `defer user.Cleanup()` | `defer user.Cleanup(t)` |
| `resp, _ := client.Get(...)` | `resp, err := ...; require.NoError(t, err)` |
| Missing `defer resp.Body.Close()` | Add after error check |
| `func testLogin(t *testing.T)` | `func TestLogin_Success(t *testing.T)` |

### D. Special Patterns

**OAuth Token Flow**:
```go
user := testutils.CompleteSignup(t, baseURL)
defer user.Cleanup(t)

// 1. Get authorization code
resp, err := user.Client.Get(baseURL + "/authorize?response_type=code&client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}")
require.NoError(t, err)
defer resp.Body.Close()

location := resp.Header.Get("Location")
code := extractCodeFromURL(location)

// 2. Exchange for token
client := testutils.NewTestClient()
form := url.Values{
    "grant_type": {"authorization_code"},
    "code":       {code},
    "client_id":  {"{CLIENT_ID}"},
    "redirect_uri": {"{REDIRECT_URI}"},
}
resp, err = client.PostForm(baseURL+"/api/v1/auth/token", form)
require.NoError(t, err)
defer resp.Body.Close()
```

**Verification Code Flow**:
```go
email := testutils.GenerateTestEmail()
client := testutils.NewTestClient()

logOffset, _ := testutils.GetServerLogOffset()
resp := testutils.DoSignupBegin(t, client, baseURL, email, "Test1234")
defer resp.Body.Close()

code, err := testutils.GetVerificationCodeFromServerLogAfterOffset(t, email, 15*time.Second, logOffset)
require.NoError(t, err)

resp = testutils.DoSignupVerify(t, client, baseURL, code)
defer resp.Body.Close()

defer testutils.CleanupUser(t, email)
```

---
