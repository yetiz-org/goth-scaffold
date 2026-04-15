---
name: filterable-handlers
description: Use when adding RSQL filter, sort, or search capability to a handler list method (Index), including designing the HandlerSchema, wiring queryfilter.FromRequest + ToOption, selecting filter/sort fields, and writing schema unit tests.
---

# Filterable Handlers

Add RSQL filter and sort capability to a handler list method using the `app/components/queryfilter` package.

## Package API

| Function           | Signature                                                 | Purpose                                                    |
|--------------------|-----------------------------------------------------------|------------------------------------------------------------|
| `FromRequest`      | `(req) → (Node, []SortField, error)`                      | Parse `q=` (RSQL filter) and `s=` (sort) from HTTP request |
| `ToOption[T]`      | `(node, sorts, schema) → (DatabaseQueryOption[T], error)` | Convert parsed AST + sorts into a repository option        |
| `ParseFilter`      | `(q string) → (Node, error)`                              | Parse RSQL string (for tests)                              |
| `ParseSort`        | `(s string) → []SortField`                                | Parse sort string like `+name,-score`                      |
| `ValidateNode[T]`  | `(node, schema) → error`                                  | Validate filter AST against schema (for tests)             |
| `ValidateSorts[T]` | `(sorts, schema) → error`                                 | Validate sort fields against schema (for tests)            |

## Schema Field Definition

`Schema[T]` is `map[string]FieldDef[T]`. Each key is the query field name clients use in `q=` or `s=`.

```go
type FieldDef[T any] struct {
Type       FieldType // FieldTypeString | FieldTypeInt | FieldTypeBool | FieldTypeFloat
AllowedOps []Op      // if empty → all ops accepted

// Filter — first non-nil wins:
Column   string         // Level 1: direct DB column
SQLExpr  func (op Op, val string) (sql string, args []any, err error) // Level 2: custom SQL
FilterFn func (row T, op Op, val string) bool // Level 3: in-memory

// Sort — independent of filter:
SortColumn string
SortFn     func (a, b T) int
}
```

**Mixed sort rule:** if ANY sort field has `SortFn`, ALL sorting is done in-memory. Prefer `SortColumn` whenever
possible.

## Step 1 — Analyze the Handler

1. Identify the target `Index` (or list) method.
2. Read the model struct being queried.
3. Read the response struct — only expose fields visible in the response.
4. Note any existing `.Where(...)` or `.Order(...)` — schema conditions add to these.

## Step 2 — Wire the Handler

```go
func (h *Foo) Index(...) ghttp.ErrorResponse {
node, sorts, err := queryfilter.FromRequest(req)
if err != nil {
return erresponse.InvalidRequestWrongBodyFormat
}

opt, err := queryfilter.ToOption[*models.Foo](node, sorts, h._fooSchema())
if err != nil {
return erresponse.InvalidRequestWrongBodyFormat
}

items, _ := h._deps().FooRepo.Find(opt)
resp.JsonResponse(items)
return nil
}
```

## Step 3 — Define the Schema

```go
// _fooSchema returns the filter/sort schema for Foo list endpoints.
// Expose only fields visible in FooGetResponse.
func (h *Foo) _fooSchema() queryfilter.Schema[*models.Foo] {
return queryfilter.Schema[*models.Foo]{
"name": {
Type:       queryfilter.FieldTypeString,
Column:     "name",
SortColumn: "name",
},
"status": {
Type:       queryfilter.FieldTypeString,
AllowedOps: []queryfilter.Op{queryfilter.OpEq},
Column:     "status",
},
"created_at": {
Type:       queryfilter.FieldTypeInt,
SortColumn: "created_at",
},
"active": {
Type: queryfilter.FieldTypeBool,
SQLExpr: func (op queryfilter.Op, val string) (string, []any, error) {
switch val {
case "true":
return "deleted_at IS NULL", nil, nil
case "false":
return "deleted_at IS NOT NULL", nil, nil
default:
return "", nil, fmt.Errorf("invalid value: %s", val)
}
},
},
}
}
```

Schema method naming: `_<resourceName>Schema()` — underscore prefix, PascalCase.

## Step 4 — Write Schema Tests

Test the schema separately from the handler. Cover:

1. **Valid filter positive** — valid fields and ops pass `ValidateNode`.
2. **Valid filter negative** — unknown fields, disallowed ops fail.
3. **Valid sort positive** — sortable fields pass `ValidateSorts`.
4. **Valid sort negative** — filter-only fields (no `SortColumn`) fail sort validation.
5. **SQLExpr edges** — boundary values for custom SQL functions.

```go
func TestFooSchema_ValidFilter_Positive(t *testing.T) {
t.Parallel()
schema := (&Foo{})._fooSchema()
cases := []struct{ name, q string }{
{"by name", `name=="acme"`},
{"by status", `status=="active"`},
}
for _, tc := range cases {
t.Run(tc.name, func (t *testing.T) {
t.Parallel()
node, err := queryfilter.ParseFilter(tc.q)
if err != nil {
t.Fatalf("ParseFilter error: %v", err)
}
if err := queryfilter.ValidateNode(node, schema); err != nil {
t.Errorf("unexpected error: %v", err)
}
})
}
}
```

## Step 5 — Verify

```bash
go build ./...
go test -v -count=1 ./tests/units/...
```

## Common Mistakes

| Mistake                               | Fix                                                      |
|---------------------------------------|----------------------------------------------------------|
| Exposing all model columns            | Only expose fields visible in the response               |
| `OpLike` for enum fields              | Enums use `AllowedOps: [OpEq]` only                      |
| Missing `FieldType` for int/bool      | Set correct type or parser defaults to string            |
| `SortFn` on a field with `SortColumn` | Mixed sort forces in-memory sort — use `SortColumn` only |
| Exported schema method                | Must be `_PascalCase` (underscore prefix)                |
| FK / tenant ID without auth check     | Do not expose without scoped base query                  |
