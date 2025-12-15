---
trigger: model_decision
description: Model unit tests guidelines (pure logic, no DB)
globs:
---

# Model Testing Guidelines

## 📋 OVERVIEW

Model tests are used to verify **pure data structure behavior** and **pure logic methods** within `app/models/*`.
These tests should be very fast, repeatable, and have no external side effects.

### Model Test Positioning
- **Model Unit Test**: Tests `struct` methods, field derivation logic, string assembly, encryption/encoding wrapper behavior (but not external systems).
- **Repository Integration Test**: Tests real MySQL/GORM read/write and query conditions (follow `repo-testing.md`).
- **E2E Test**: Tests cross-layer flows (HTTP + session + DB + external services) (follow `e2e-testing.md`).

### Core Principles
1. **No DB**: Model unit tests should not connect to MySQL directly, should not call repositories.
2. **No network**: Should not make HTTP calls, should not depend on external services.
3. **Deterministic**: Same input must produce same output.
4. **Isolated**: Avoid modifying global state; if unavoidable, must save/restore.
5. **No useless tests**: Don't just test "getter returns field" with no business significance; exception is "permission checks" and similar semantic APIs.

---

## 📁 FILE STRUCTURE

### Suggested Location
- `tests/units/models/*.go`: Recommended to centralize model unit tests (this folder may not exist yet).

### Package Choice
- **Prefer** `package models_test`: Test public API from external perspective, avoid depending on internals.
- **Use** `package models`: Only when testing unexported helpers or must access package-private state.

### Package Doc Comment (RECOMMENDED)
```go
/*
Package models - {Model Name} Unit Tests

This file contains unit tests for {Model Name}.
Tests cover pure logic only (no DB, no network).

Features:
  - {feature 1}
  - {feature 2}

Test Case List:
  - Test{Model}_{Method}_{Scenario}: {brief description}
*/
package models_test
```

---

## 🏷️ NAMING CONVENTIONS

### Test Function Names
**Pattern**: `Test{Model}_{Method}_{Scenario}`

Examples:
```go
TestUser_AvatarURLWithSize_NoAvatar
TestUser_AvatarURLWithSize_UnknownSizeDefaultsTo480
TestSitePermission_EncryptedUserID_RoundTrip
```

---

## 🧪 WHAT TO TEST (SCOPE)

### 1. TableName() (GORM mapping)
Although the logic is thin, getting it wrong directly causes DB mapping failure, **allowed to test**.

### 2. Derived / Computed Methods (High Priority)
For models like `users.go` / `site_permissions.go`:

#### User
- `AvatarURLWithSize(size)`
  - `Avatar == nil` → `""`
  - `*Avatar == ""` → `""`
  - `size == original/960/480` → correct corresponding suffix
  - `size == unknown` → fallback to `480`
- `AvatarURL()`/`AvatarURLOriginal()`/`AvatarURL960()`
  - Verify correct default size behavior when calling `AvatarURLWithSize`

#### SitePermission
- `EncryptedUserID()` / `EncryptedSitePermissionID()`
  - **Round-trip test**: Encrypted result must not be empty, and can be restored using corresponding decrypt.

### 3. Semantic Predicate Methods (Can test, but needs justification)
For example `IsAdmin()` - even if it just returns a field, it's often used as a business API.

Recommendations:
- Only test when it's a "semantic API that external code depends on".
- Don't write repetitive meaningless tests for every bool.

---

## 🚫 WHAT NOT TO TEST

- Don't test Repository CRUD (put in `tests/units/repositories`, follow repo-testing).
- Don't test GORM behavior (belongsTo / preload / migration) itself.
- Don't test HTTP handlers, acceptances, session/cookie (put in E2E/handler tests).
- Don't depend on `daemons/*` startup process in model unit tests.

---

## 🔒 GLOBAL STATE & PARALLEL EXECUTION

Model tests are usually very suitable for parallel execution, but **prerequisite is no shared mutable state**.

### Rule of Thumb
- **Pure functions / no global mutation** → Can use `t.Parallel()`.
- **Needs global mutation** (e.g., `conf.Config()` singleton, `crypto.KeyPrefix`) → **DO NOT** `t.Parallel()`, and must restore state.

### Examples of global state in this repo
- `conf.Config()`: Global singleton, some methods may depend on CDN domain or other config settings
- `crypto.KeyPrefix`: Global variable, affects `Encrypt*` / `Decrypt*`

### Safe pattern (save & restore)
```go
old := conf.Config().Credentials.CDNDomainName
conf.Config().Credentials.CDNDomainName = "cdn.example.com"

t.Cleanup(func() {
    conf.Config().Credentials.CDNDomainName = old
})
```

---

## ✅ ASSERTION RULES

- Use `require` for preconditions (nil, error, necessary setup)
- Use `assert` for result verification
- Every assertion should have a clear message (English preferred)

---

## 🧩 TEST CASE DESIGN CATEGORIES

- **Happy Path**: Normal input → Normal output
- **Boundary**: nil/empty/unknown enum/invalid size
- **Error Handling**: Wrapper method returns empty string or default value on error
- **Compatibility**: e.g., unknown size fallback behavior must not break

---

## 🧪 RECOMMENDED COMMANDS

```bash
# Run model unit tests only (suggested)
go test -v ./tests/units/models/...

# Run all tests (final verification)
go test -v ./...
```

---

## 📋 CHECKLIST BEFORE SUBMISSION

- [ ] No DB / no network
- [ ] Deterministic outputs
- [ ] No global state leak (save & restore)
- [ ] `require` vs `assert` usage correct
- [ ] No compile errors
- [ ] Tests pass
