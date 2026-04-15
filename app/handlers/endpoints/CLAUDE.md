# app/handlers/endpoints

> For file layout, naming conventions, and section ordering, see [handlers/CLAUDE.md](../CLAUDE.md).

## Core Contract

- Embed the shared handler task (or a derived task) in every endpoint handler struct.
- Implement HTTP behavior by overriding struct methods: `Index`, `Get`, `Post`, `Put`, `Patch`, `Delete`.
- Do not create package-level handler functions — behavior lives on the struct.
- Expose one package-level singleton per handler: `var Handler<Name> = &<Name>{}`.

## Dispatch Lifecycle

`ghttp.DispatchHandler` calls in this order on every request:

1. `PreCheck(req, resp, params)`
2. `Before(req, resp, params)`
3. Method dispatch (see table below)
4. `After(req, resp, params)`
5. `ErrorCaught(req, resp, params, err)` — only on non-nil error return

## Method Dispatch Table

| HTTP Verb | Tries first | Falls back to                                 | Falls back to |
|-----------|-------------|-----------------------------------------------|---------------|
| `GET`     | `Index`     | `Get` (if `Index` returns `NotImplemented`)   | 405           |
| `POST`    | `Create`    | `Post` (if `Create` returns `NotImplemented`) | 405           |
| `PUT`     | `Put`       | —                                             | 405           |
| `PATCH`   | `Patch`     | —                                             | 405           |
| `DELETE`  | `Delete`    | —                                             | 405           |
| `OPTIONS` | `Options`   | —                                             | 405           |

**Key implication**: `Index` and `Create` are opt-in override points. If you don't override
them, the framework falls back silently — no boilerplate needed. Use `Index` for list endpoints
(`GET /v1/foos`) and `Get` for single-item endpoints (`GET /v1/foos/:id`).

## Register() Contract

`Register()` is called once at startup for idempotent initialization (e.g. wiring sub-routes,
loading config). Rules:

- Never call `Register()` from an HTTP method handler.
- `Register()` must be safe to call multiple times (idempotent).
- Omit `Register()` entirely if there is nothing to initialize.

## Error Response Contract

All HTTP method handlers return `ghttp.ErrorResponse`.

| Situation                           | Return                                        |
|-------------------------------------|-----------------------------------------------|
| Success with body                   | `nil` — framework sends HTTP 200 + JSON body  |
| Success, no body                    | `nil` — framework sends HTTP 200 + empty body |
| Not found                           | `ghttp.NotFound`                              |
| Unauthorized                        | `ghttp.Unauthorized`                          |
| Bad input                           | `ghttp.BadRequest`                            |
| Not implemented (triggers fallback) | `ghttp.NotImplemented`                        |

**Never return `{"success": true}`** — HTTP 200 + nil body is the success signal.

## Inherited Helpers

| Helper                | Purpose                                                       |
|-----------------------|---------------------------------------------------------------|
| `GetNode(params)`     | Current `ghttp.RouteNode`                                     |
| `GetNodeId(params)`   | Resolves current node identifier                              |
| `Redirect(url, resp)` | Shared redirect wrapper                                       |
| `T(message, lang)`    | Translate message key                                         |
| `Lang(req)`           | Resolve language: query → session → Accept-Language → default |

## Logging

```
handlers:HandlerName.Method#section!action
```

English only. Log at entry and on all error paths.

## Never

- Read framework-private param keys directly when an inherited helper covers the case.
- Embed concrete handlers that already own HTTP methods (only embed approved base tasks).
- Call `Register()` from inside an HTTP method handler.
