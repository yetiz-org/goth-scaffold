---
name: api-to-docs
description: Use when OpenAPI paths, methods, schemas, enums, or operationIds may have drifted from route.go, handler implementations, acceptance chains, or actual response output.
---

# API to Docs

## Core Rule

Document the API the code serves today — not what old YAML claims.
The handler implementation is the truth; OpenAPI is the output.

## When to Use

- A handler or route changed under `app/handlers/endpoints/**` or `app/handlers/route.go`.
- Request/response structs changed.
- Response serialization logic changed.
- Enum, type, nullability, or security drift is suspected.
- `operationId` duplicates or stale paths need cleanup.

Don't use when:
- The change is a pure HTML/page route with no OpenAPI contract.
- The endpoint is explicitly out of current spec scope.

## Source of Truth (in order)

1. `app/handlers/route.go` — endpoint families and handler wiring.
2. Route-level and group-level acceptances — auth and permission requirements.
3. Handler method implementations — supported HTTP methods and their behavior.
4. Request parsing code — path params, query params, form fields, JSON body.
5. Actual response serialization — `resp.JsonResponse(...)`, `resp.SetBody(...)`, binary writers.
6. Constants, validation tags, SQL schema for enums and examples.
7. Existing OpenAPI only as a **diff target** — never as truth.

## Workflow

### 1. Lock scope

- Stay within the named endpoints or files.
- Trace from handler → route → acceptance chain for auth requirements.

### 2. Build effective path and method matrix

For each in-scope endpoint:
- Effective HTTP path (may differ from literal `SetEndpoint` string if params are injected by helpers)
- Supported methods (implemented, not just declared)
- Auth: token type, permission level required

### 3. Resolve request shape

Inspect actual parsing code:

| Source | Pattern |
|--------|---------|
| Path param | `params["name"]` or helper accessor |
| Query param | `req.URL.Query().Get("name")` |
| Form field | `req.FormValue("name")` |
| JSON body | struct fields in `json.Unmarshal` target |

### 4. Resolve response shape

Follow the exact response-writing call:

```go
resp.JsonResponse(payload)          // → payload fields are the schema
resp.SetBody(bytes, contentType)    // → binary or non-JSON response
return nil                          // → 200 with no body
```

Do not derive schema from model structs unless the handler serializes one-to-one — this is rarely true.

### 5. Align security

From the acceptance chain:
- Whether auth is required (token, session, or none)
- Whether permission checks (admin, owner, etc.) apply

### 6. Type and nullability mapping

| Wire behavior | OpenAPI |
|---------------|---------|
| Unix timestamp integer | `type: integer, format: int64` |
| RFC3339 string | `type: string, format: date-time` |
| Pointer field that becomes `null` | `nullable: true` |
| Pointer field omitted via `omitempty` | optional (not nullable) |
| Enum | list only values from code constants, not invented |

### 7. Verify

```bash
# YAML syntax check
ruby -ryaml -e "YAML.safe_load(File.read('docs/openapi/openapi.yaml'))" 2>/dev/null || \
python3 -c "import yaml,sys; yaml.safe_load(open('docs/openapi/openapi.yaml'))"

# Duplicate operationId check
grep 'operationId:' docs/openapi/openapi.yaml | sort | uniq -d

# Build to ensure no code was broken
go build ./...
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Copying model fields into response | Follow response-writing call site |
| Treating all pointer fields as nullable | Check `omitempty` behavior |
| Adding `{"success": true}` for 200 | Document only what is actually returned |
| Keeping yaml-only fields | Remove stale YAML fields, not just add missing ones |
| Taking `SetEndpoint` path literally | Check if helpers inject implicit path params |
| Old YAML as truth | Re-derive from current code |

## Final Check

Before claiming complete:
1. Re-read changed handler files.
2. Re-read route and acceptance wiring.
3. List every documented field — verify code has it.
4. Check both directions: code-only fields AND yaml-only fields.
