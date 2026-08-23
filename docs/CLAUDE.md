# Documentation Rules

## Purpose

These rules apply under `docs/`. The current documentation surface is the static landing page and the OpenAPI viewers/specification under `docs/openapi/`.

Keep scaffold documentation generic. Do not add product, customer, vendor, private host, real credential, or business-workflow examples.

## Sources of Truth

For OpenAPI work, code served today is authoritative in this order:

1. `app/handlers/route.go` for effective route families and wiring.
2. Parent and child acceptance chains for authentication and authorization.
3. Endpoint methods for supported HTTP methods and behavior.
4. Request parsing for path, query, form, and JSON inputs.
5. Actual response serialization for status, body shape, headers, and content type.
6. Constants and validation rules for enums, formats, and limits.
7. `docs/openapi/goai.yaml` for generator metadata, profiles, output names, and exclusions only.
8. `docs/openapi/openapi.yaml` as generated output and the diff target, never as behavioral truth.

When paths, methods, request/response shapes, enums, security, or `operationId` values may drift, invoke the repository's `api-to-docs` skill and follow its scope/verification contract.

## Operating Loop

| Trigger | Action | Check |
| --- | --- | --- |
| A route, acceptance, handler, or wire struct changes | Trace the effective request and response contract before editing OpenAPI. | Every documented method, field, status, and security rule exists in current code. |
| OpenAPI changes | Update namespaced `@goai.*` handler directives or generator config, then run `make docs-build`. Remove stale YAML-only claims. | Generated YAML parses, matches served behavior, and duplicate `operationId` output is empty. |
| Viewer HTML changes | Preserve local spec loading, OAuth redirect behavior, external asset integrity, and both Swagger/Redoc entry points. | Open the changed page through a local HTTP server and inspect browser console/network errors. |
| A command/path/example is added | Verify it exists in this repository and contains no private environment data. | The command/path resolves and examples remain neutral. |
| A standalone implementation plan is requested | Use the available `writing-plans` workflow. Do not invent a `docs/plans` contract from another project. | The plan package passes that workflow's current quality gate. |

## OpenAPI Contract

- Keep OpenAPI at the version already declared unless a migration is explicitly requested.
- Derive schemas from serialized handler output, not model structs unless serialization is proven one-to-one.
- Distinguish optional fields from nullable fields using actual encoding behavior.
- List only enum values present in current constants or validation.
- Preserve the project convention that JSON bodies use a top-level object and empty successes have no synthetic success body.
- Do not document an endpoint, parameter, status, or response field that code does not serve.

## Validation

After changing GoAI directives, generator config, or `docs/openapi/openapi.yaml`:

```bash
make docs-build
ruby -ryaml -e "YAML.safe_load(File.read('docs/openapi/openapi.yaml'))" 2>/dev/null || python3 -c "import yaml; yaml.safe_load(open('docs/openapi/openapi.yaml'))"
awk '/operationId:/{print $2}' docs/openapi/openapi.yaml | sort | uniq -d
git diff --check -- docs/
```

The duplicate command must print nothing. If the API code also changed, run `go build ./...` and the code gates from the root rules.

After changing viewer HTML, serve `docs/` over HTTP instead of opening `file://` directly, then verify the changed page loads its selected local YAML and has no console/network error.

For rule-only changes, run the rule validator and final diff/status inspection; do not run Go tests solely for Markdown.
