# app/handlers

## Route Registration

All routes are registered in `app/handlers/route.go`.
Child routes automatically inherit parent acceptances — do not duplicate them on child routes.

## Directory Layout

```
app/handlers/
├── route.go              # all route wiring
├── service.go            # HTTP server lifecycle (Start / Stop)
├── initializer.go        # channel pipeline bootstrap (MIME, headers)
├── track_handler.go      # base channel handler with HTTP context helpers
├── endpoints/            # one file per resource (see endpoints/CLAUDE.md for rules)
│   ├── handlertask.go    # base task with shared defaults and helpers
│   └── v1/
│       └── foos.go
├── acceptances/          # request validation middleware
└── minortasks/           # shared pre-handler tasks (auth, param injection, etc.)
```

## Acceptances

Request validation middleware lives in `app/handlers/acceptances/`.
Register per-route or per-group in `route.go`.
Group-level acceptances are inherited; do not re-register them on child endpoints.

Scaffold-provided acceptances:

| File                                | Var               | Purpose                                                |
|-------------------------------------|-------------------|--------------------------------------------------------|
| `check_basic_auth.go`               | `HCheckBasicAuth` | Validates HTTP Basic Auth against config               |
| `skip_method_options_acceptance.go` | —                 | Skips validation for OPTIONS requests (CORS preflight) |

## Minortasks

Reusable pre-handler tasks live in `app/handlers/minortasks/`.
Register on a route group in `route.go` to inject context (e.g. decoded token) before the endpoint handler runs.

Scaffold-provided minortasks:

| File                     | Var                   | Purpose                                                       |
|--------------------------|-----------------------|---------------------------------------------------------------|
| `decode_access_token.go` | `TaskDecodeSiteToken` | Decodes Bearer token and injects identity into request params |

Add new minortasks when multiple endpoint groups share the same pre-processing logic.

## Handler File Layout

Every endpoint file has exactly **three sections**, in this order:

```go
// =============================================================================
// Struct and Init
// =============================================================================

// Section 1 — struct type, package-level singleton, Register().
// Public identifiers first (HandlerFoo, Register), private last.

type Foo struct {
	endpoints.HandlerTask
	_someService services.SomeService
}

var HandlerFoo = &Foo{}

func (h *Foo) Register() { ... }

// =============================================================================
// Request / Response Types
// =============================================================================

// Section 2 — all request/response payload structs for this handler.

type FooGetResponse struct { ... }
type FooPatchRequest struct { ... }

// =============================================================================
// HTTP Handlers
// =============================================================================

// Section 3 — HTTP method handlers in canonical order, then private helpers.
// Public HTTP methods come before private (_PascalCase) helpers.

func (h *Foo) Index(...) ghttp.ErrorResponse { ... }  // GET  list
func (h *Foo) Get(...)   ghttp.ErrorResponse { ... }  // GET  single
func (h *Foo) Head(...)  ghttp.ErrorResponse { ... }  // HEAD single
func (h *Foo) Post(...)  ghttp.ErrorResponse { ... }  // POST
func (h *Foo) Put(...)   ghttp.ErrorResponse { ... }  // PUT
func (h *Foo) Patch(...) ghttp.ErrorResponse { ... }  // PATCH
func (h *Foo) Delete(...) ghttp.ErrorResponse { ... } // DELETE

func (h *Foo) _buildQuery(...) { ... }  // private helper — after all HTTP methods
```

### Ordering Rules

- **Section 1**: public first — `var HandlerFoo` and `Register()` before any private fields.
- **Section 3 — HTTP method order**: `Index` → `Get` → `Head` → `Post` → `Put` → `Patch` → `Delete`.
- **Private helpers** (`_PascalCase`) go after all HTTP method overrides within section 3.
- Do not add, reorder, or remove sections.

### Override Rules

- Override only the methods the endpoint actually handles.
- Unoverridden methods use task defaults (most return 405 Method Not Allowed).
- Default `Index` returns `ghttp.NotImplemented` — enables automatic GET fallback to `Get`.
- Default `Head` returns `nil` (200 OK with empty body) — HEAD does **not** fall back to `Get`; override only to deny HEAD or emit custom headers.
- Use `Index` for list/root (`GET /v1/foos`); use `Get` for single-item (`GET /v1/foos/:id`).
- Use `Head` only when the endpoint must emit response headers without a body (e.g. cache validation, length probing); otherwise leave unoverridden.
- Use `Patch` for partial updates instead of overloading `Post`.

## Naming

```go
type Foo struct { endpoints.HandlerTask }
var HandlerFoo = &Foo{}
```

Payload types: `<Handler><Method>Request` / `<Handler><Method>Response`

```go
type FooGetResponse struct { Name string `json:"name"` }
type FooPatchRequest struct { Name *string `json:"name"` }
```

- Use `*T` fields in request structs when partial update must distinguish missing vs zero.

## Logging

```
handlers:HandlerName.Method#section!action
```

English only. Log at entry and on error paths.

## Detailed Rules

See `endpoints/CLAUDE.md` for:

- Method dispatch semantics (PreCheck → Before → Method → After → ErrorCaught)
- Inherited helpers (GetNode, GetNodeId, Redirect, T, Lang)
- Complete example with Tables and parameter documentation
