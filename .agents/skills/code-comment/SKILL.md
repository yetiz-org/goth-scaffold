---
name: code-comment
description: Use when Go source files are missing standardized comments, or when implementation changes may have caused existing comments to drift from actual behavior. Applies to handlers, services, repositories, models, workers, and any exported Go code.
---

# Code Commenting for Go

## Core Rule

Document what the code actually does today. Reconstruct from code evidence — never copy old comments verbatim or invent behavior.

## When to Use

- An exported function, method, type, or constant is missing a doc comment.
- Implementation changed but comments were not updated.
- A handler file is missing its file-level header comment.

Do not use when:
- The target is unexported and the logic is self-evident from naming.
- The change is unrelated to documentation.

## Source of Truth

1. The actual implementation — read it, don't assume.
2. Route wiring (`app/handlers/route.go`) for HTTP paths.
3. Request parsing code for parameter documentation.
4. Service/repository calls for Tables annotations.

Never trust old comments as truth.

## Workflow

### 1. Lock scope

```bash
# Find all Go files in scope
rg --files <target_dir> -g '*.go' -g '!*_test.go' | sort
```

### 2. For each file in scope

1. Check for file-level header comment (before `package`).
2. Find all exported identifiers:
   ```bash
   rg -n "^(func|type|var|const) [A-Z]" <file>
   ```
3. For each exported identifier, verify a doc comment exists and matches current behavior.
4. If missing or stale — reconstruct from current implementation.

### 3. Verify compilation

```bash
go build ./...
```

Fix all compilation errors before completion.

## Comment Formats

### File Header (before `package`)

When the file directly touches DB tables:

```go
// Package handlers implements the user management endpoints.
// Tables:
//   - users
//   - user_sessions

package handlers
```

If no direct DB access is relevant, omit the Tables block:

```go
// Package conf loads and validates application configuration.

package conf
```

### Exported Function / Method

```go
// FindByEmail returns the user with the given email address, or nil if not found.
// Errors are logged internally; callers receive nil on failure.
func (r *UserRepository) FindByEmail(email string) *User {
```

### HTTP Handler Method

```go
// Get GET /api/v1/users/:id
// Returns the user identified by :id.
//
// Path:
//   - id: encrypted user ID — required
//
// Tables:
//   - users
func (h *Users) Get(...) ghttp.ErrorResponse {
```

HTTP method comment structure:
- First line: `// Method HTTP_VERB /path`
- Second line: one-sentence description
- Optional sections: `Path`, `Query`, `Form`, `Request body (JSON)` — include only sections that exist in the implementation
- Optional `Tables:` block — only when handler directly reads/writes DB tables

### Type / Struct

```go
// SiteSetting represents a runtime-configurable key/value pair grouped by category.
// Use SiteSettingRepository to read and write instances.
type SiteSetting struct {
```

### Interface

```go
// SiteSettingRepository defines data access for SiteSetting records.
type SiteSettingRepository interface {
```

## Parameter Detection

Extract from implementation:

| Source | Pattern |
|--------|---------|
| Path param | `params["name"]` |
| Query param | `req.URL.Query().Get("name")` |
| Form field | `req.FormValue("name")` |
| JSON body | struct fields in `json.Unmarshal` target |

Document only parameters actually read by the implementation.

## HTTP Method Mapping

| Method name | HTTP verb |
|-------------|-----------|
| `Index` | `GET` (list) |
| `Get` | `GET` (single) |
| `Post` | `POST` |
| `Patch` | `PATCH` |
| `Put` | `PUT` |
| `Delete` | `DELETE` |

## Completion Checklist

- [ ] Every exported identifier has a doc comment.
- [ ] HTTP handler methods follow the standard format.
- [ ] Tables blocks use bullet list format and accurate table names.
- [ ] Parameter sections match actual parsing code.
- [ ] `go build ./...` passes.

## Common Mistakes

| Mistake | Correct action |
|---------|---------------|
| Copying old comment text | Rebuild from current code |
| Guessing HTTP path from filename | Read route.go |
| Listing parameters not in code | Document only values actually read |
| Adding Tables block to every file | Add only when relevant |
