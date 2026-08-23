# Service Layer Rules

## Purpose

These rules apply to `app/services/**`. Services own application workflows and transaction coordination. Handlers own transport, repositories own persistence, models own data shape, components own reusable infrastructure, and workers own background execution.

Examples and new service names must stay generic in this scaffold. Do not introduce product, customer, vendor, dataset, or business-workflow names.

## Priorities

1. Preserve workflow invariants, transaction boundaries, failure semantics, and public behavior.
2. Keep dependencies explicit, lazy where connector access is optional, and replaceable in tests.
3. Keep one owner for each transformation or validation rule.
4. Prefer direct code over compatibility wrappers or speculative abstractions.

## Operating Loop

| Trigger | Action | Check |
| --- | --- | --- |
| Before changing a service | Read the complete service, dependency container, repository interfaces, direct callers, non-call references, tests, and ancestor rules. | The affected callers and preserved contracts are known before the edit. |
| Before adding behavior | Search services, components, helpers, models, and repositories for the existing owner. | Reuse or extend one owner; do not duplicate the rule. |
| Before a multi-row or multi-table mutation | Select one transaction owner and pass the same `*gorm.DB` to every `Tx` repository/service call. | No nested transaction or non-transactional write escapes the operation. |
| Before using a connector | Follow the connector's `Enabled()` and lazy-access rules. | Package initialization opens no connection and optional connectors fail safely. |
| After a dependency or workflow change | Compile, run focused existing tests, then run the repository worktree gate. | Commands exit `0` and the final diff contains no unrelated change. |

## Dependencies and Construction

- Use `app/services/example/service.go` as the local structural reference, not as domain content to copy.
- Keep repositories and sibling services behind private fields or a package-private `_Deps` value.
- Use `_PascalCase` for private fields and methods. Use `example._Deps` as the package-default repository pattern.
- Keep only dependencies the service actually calls. Do not mirror raw reader, writer, or session accessors in `_Deps` when repository factories already own them.
- Pass `database.Reader`, `database.Writer`, or keyspace access to repository factories as functions when lookup must remain lazy. Do not resolve optional connections during package initialization.
- Preserve zero-value behavior when safe. Add a constructor only when it selects established defaults, validates a real invariant, or owns a lifecycle/resource.
- Avoid process-wide mutable service singletons. If shared state is required, name its owner and protect it with mutex, atomic, channel, or immutable initialization.

## Methods and Transactions

- Exported methods document purpose, input constraints, return values, nil/empty behavior, errors, and side effects. The comment must start with the symbol name and match current code.
- Use named results where they clarify the project contract; do not add names that only repeat obvious types.
- A method that joins an external GORM transaction takes `tx *gorm.DB` first and uses a `Tx` suffix.
- A `Tx` method must use the supplied transaction for every SQL operation in that workflow. It must not silently fall back to a package writer.
- Keep validation and transformation in the service only when the workflow owns it. Persistence query construction stays in repositories.
- Preserve error identity when callers use `errors.Is` or `errors.As`. Log at the owning boundary with `<service-package>:Service.Method#section!action`; do not log the same error at every layer.
- Return nil/empty/error indicators consistently with the repository interface and the documented service contract. Do not convert an unexpected failure into apparent success.

## Package Shape

- Put each cohesive service capability in `app/services/<capability>/` with package-local `Service` and `_ServiceDeps` names. The neutral reference is `example.Service`.
- Keep `app/services/` free of root Go declarations and forwarding shims. Callers import the owning capability package directly.
- Concrete sibling-service dependencies use named private fields. Do not use anonymous embedding or promoted calls that hide the dependency owner.
- Public package moves or type renames require caller/reference inventory and explicit compatibility approval.

## Safety Boundaries

| Trigger | Required action |
| --- | --- |
| Context-aware I/O | Propagate `context.Context`, respect cancellation, and release resources with `defer`. |
| External service, filesystem, or database work | Set bounded timeouts, validate inputs, close resources, and preserve retry/idempotency rules. |
| Sensitive data | Never log credentials, tokens, personal data, or raw secrets. |
| Concurrent or lazy mutable state | Identify the owner and synchronization primitive; run race detection on affected tests. |
| Test dependency injection | Replace package-private `_Deps` fields or reuse existing test helpers; restore shared state with `t.Cleanup`. |

## Validation Commands

Run from the repository root and select focused tests for the affected service first.

```bash
gofmt -w <changed-go-files>
gofmt -d <changed-go-files>
go list ./app/services/...
go test -v -count=1 -run '^$' ./app/... ./tests/units/...
make local-test
make worktree-test
git diff --check
git status --short
```

For a rule-only edit, run the rule validator and Git diff/status checks instead of Go tests.
