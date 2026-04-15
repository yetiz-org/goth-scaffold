---
name: writing-handlers
description: Use when creating, updating, splitting, or refactoring HTTP handlers or REST endpoints under app/handlers/endpoints/**, or when route wiring, acceptances, minortasks, ghttp request parsing, response shape, or pagination may change.
---

# Writing Handlers

## Core Rule

Read the route chain and existing handler contract before changing any code.
Route wiring, acceptance inheritance, ghttp parameter resolution, response shape, and handler boundary ownership are all part of the contract.

## When to Use

- Add a new handler under `app/handlers/endpoints/**`.
- Update an existing HTTP method (`Index`, `Get`, `Post`, `Patch`, `Put`, `Delete`).
- Add or change `SetEndpoint(...)` / `SetGroup(...)` in `app/handlers/route.go`.
- Change auth, permission, or user-loading behavior around a handler.
- Change response shape, pagination, path params, or JSON body handling.
- Split or move handler files.
- Update handler tests under `tests/units/handlers/` or `tests/e2e/`.

Don't use when:
- Work is purely OpenAPI doc sync with no code changes.
- Work is only worker/cron logic with no HTTP handler impact.

## Source of Truth (read in order)

1. `app/handlers/route.go` — endpoint families, package ownership, inherited acceptances.
2. `app/handlers/endpoints/handlertask.go` — default HTTP behavior and helper accessors.
3. `app/handlers/acceptances/` — request validation middleware.
4. `app/handlers/minortasks/` — shared pre-handler tasks.
5. Nearby handlers in the same package + existing tests in `tests/units/handlers/` and `tests/e2e/`.

## Workflow

### 1. Lock scope

- Identify the concrete endpoint path, package, and resource.
- Check if a handler already exists in the target folder before creating one.
- Map API paths back to `route.go` first.

### 2. Trace route + acceptance chain

- Find the relevant `SetGroup(...)` and `SetEndpoint(...)` entries.
- Record inherited acceptances and minortasks.
- Don't duplicate parent acceptances on child routes — inheritance applies automatically.

### 3. Define the handler struct

```go
// =============================================================================
// Struct and Init
// =============================================================================

type Foo struct {
    endpoints.HandlerTask
    _fooService services.FooService
}

var HandlerFoo = &Foo{}

func (h *Foo) Register() {
    h.Do(func() {
        h._fooService = &services.FooService{}
    })
}

// =============================================================================
// Request / Response Types
// =============================================================================

type FooPostRequest struct {
    Name string `json:"name"`
}

// =============================================================================
// HTTP Handlers
// =============================================================================

// Post POST /api/v1/foos
// Creates a new Foo resource.
func (h *Foo) Post(...) ghttp.ErrorResponse { ... }
```

Rules:
- Handler-owned services and repositories are **private** (`_camelCase` field, `_PascalCase` accessor).
- Do not expose dependencies as exported fields.
- Do not anonymously embed another concrete handler.
- Do not call `Register()` from HTTP methods.
- Keep HTTP methods above private helpers within the file.

### 4. Request validation

```go
if len(req.Body().Bytes()) == 0 {
    return erresponse.InvalidRequestWrongBodyFormat
}
body := &FooPostRequest{}
if err := json.Unmarshal(req.Body().Bytes(), body); err != nil {
    return erresponse.InvalidRequestWrongBodyFormat
}
```

Always check body length before unmarshal. Validate required fields after.

### 5. Response shape

- `resp.JsonResponse(payload)` for data responses.
- `return nil` for 200 with no body — never return `{"success": true}`.
- Return explicit empty slices when repo returns empty, not null.

### 6. Logging

```go
kklogger.InfoJ("handlers:FooHandler.Post#create!start", nil)
kklogger.ErrorJ("handlers:FooHandler.Post#create!db_error", err.Error())
```

English only. Format: `handlers:HandlerName.Method#section!action`.

### 7. Verify

```bash
go test -v ./tests/units/handlers/...
go test -v ./tests/e2e/...
go build ./...
```

Keep tests under `tests/`, not under `app/handlers/`.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Skipping `route.go` before editing | Trace `SetGroup`/`SetEndpoint` first |
| Duplicating parent acceptances on child routes | Child inherits from parent — add only extra checks |
| Exported dependency fields on handler | Use `_PascalCase` private method + `_camelCase` field |
| Embedding another concrete handler | Use helper structs that do not own HTTP methods |
| Missing body-length check before unmarshal | Add `if len(req.Body().Bytes()) == 0` first |
| `{"success": true}` on 200 responses | Return `nil` or actual payload only |
| Tests placed under `app/handlers/` | Tests live under `tests/units/handlers/` or `tests/e2e/` |
| Private helpers placed above HTTP methods | Keep HTTP methods near the top of the file |
