# Unit Test Rules

## Purpose

These rules apply under `tests/units/`. Unit tests must prove deterministic behavior without starting the application, network listeners, databases, Redis, Cassandra, queues, or other external services.

## Priorities

1. Meaningful regression or contract protection over test count.
2. Deterministic isolation over execution speed.
3. Concrete behavioral assertions over structural checks.
4. Small setup and explicit cleanup.

## Operating Loop

| Trigger | Action | Check |
| --- | --- | --- |
| Adding a unit test | Identify the exact stable behavior, regression, invariant, or validation rule. | The test name and assertions expose what would break. |
| Selecting a boundary | Exercise the smallest public or package boundary that proves the behavior. | The test does not pin a private helper or source layout without a real contract. |
| Using global config, registries, or environment | Snapshot/reset state and register cleanup before mutation. | The result is independent of order and prior tests. |
| Using a fake or mock | Model only the contract needed by the scenario. | Assertions verify observable behavior, not only calls configured by the test. |
| Finishing a change | Run the focused package, all unit tests, then the repository gate when production code changed. | Each required command exits `0`. |

## Boundaries

- Do not call `app.FlagParse`, start daemons, open real connectors, perform schema operations, or make real network requests.
- Do not add `TestMain` to bootstrap application services under `tests/units/`; move such coverage to E2E.
- Pure model/helper logic, deterministic serialization, validation, and service logic with injected dependencies belong here.
- Shared registries or singleton-like configuration may require serial tests. Call `t.Parallel()` only after proving the test and its subtests share no mutable state.
- Use `t.Setenv` for environment changes and restore other global state with `t.Cleanup` or `defer`.
- Do not use `t.Skip` to hide missing setup; a unit test with an external prerequisite is in the wrong suite.

## Assertions and Structure

- Assert exact outputs, errors, state transitions, cache behavior, or serialized results.
- A `NotNil`, `NotEmpty`, or no-panic assertion is sufficient only when that property is itself the documented contract.
- Table-driven tests are useful for multiple inputs sharing one behavior. Do not use tables to hide materially different workflows.
- Keep test-only helpers close to their consumers. Remove helpers that only wrap literals or add unused flexibility.
- When creating or substantially editing a file, add the test-file block comment required by `tests/CLAUDE.md`.

## Validation Commands

```bash
go test -v -count=1 ./tests/units/<affected-package>/...
go test -v -count=1 ./tests/units/...
# Concurrency-related changes only:
go test -v -count=1 -race ./tests/units/...
```

When production Go code or dependency metadata changed, also run `make worktree-test`. For rule-only changes, run the rule validator and diff/status checks instead.
