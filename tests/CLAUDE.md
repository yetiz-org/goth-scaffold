# Test Rules

## Purpose

These rules apply to every `go test` command and file under `tests/`. Deeper files extend this contract:

- `tests/units/CLAUDE.md` — deterministic unit-test boundaries.
- `tests/e2e/CLAUDE.md` — isolated full-stack HTTP and connector tests.

## Required Commands and Flags

Prefer the Makefile because it owns repository paths, environment variables, ports, and cleanup.

| Goal | Command |
| --- | --- |
| Unit tests | `make local-test` |
| All isolated E2E tests | `make worktree-test-e2e` |
| Full code gate | `make worktree-test` |

Every direct `go test` invocation must include `-v -count=1`. Add `-race` for concurrency-related non-E2E changes. Direct E2E runs also require `-timeout=120s` and explicit `SCAFFOLD_E2E_CONFIG`/`SCAFFOLD_E2E_BINARY` values supplied by CI or a scoped Makefile flow. The current Makefile has no focused `E2E_RUN` option; do not invent one.

Do not pipe test output through `grep`, `tail`, `/dev/null`, or another filter. Full output is required for diagnosis.

## Test Quality

| Trigger | Action | Check |
| --- | --- | --- |
| Adding or editing a test | Name the exact regression, invariant, deterministic transformation, validation rule, or public contract it protects. | Removing the test would allow a concrete bug or contract drift to pass. |
| Choosing assertions | Assert concrete output, error, state transition, serialized data, or side effect. | A stronger observable is not replaced by only `not nil`, `not empty`, `no panic`, or symbol existence. |
| Choosing setup | Use the smallest deterministic setup at the stable boundary. | The test does not bootstrap services unless it belongs under `tests/e2e`. |
| A test fails | Diagnose and fix the root cause. | Do not weaken assertions, comment out the test, or convert a regression into a skip. |
| Shared mutable test state exists | Reset it with `t.Cleanup`/`defer` and avoid unsafe parallel execution. | Order and prior tests cannot change the result. |

`t.Skip` is allowed only when the test contract explicitly makes an external dependency or local config optional. State the missing prerequisite in the skip message. It is never a way to hide a failure in an available dependency or a broken feature.

## File Documentation

When creating or substantially editing a `*_test.go` file, add a block comment before `package`:

- unit files name the behavior and source file or public boundary under test;
- E2E files name the endpoint paths and keep a test-case list synchronized with actual functions;
- comments describe observable behavior, not implementation steps or business examples.

Test names must state the behavior and scenario. Use subtests or tables only when they improve coverage without hiding setup or failure meaning.

## Coverage

Measure coverage when it helps assess the affected behavior:

```bash
go test -v -count=1 -coverprofile=coverage.out ./tests/units/...
go tool cover -func=coverage.out
```

There is no repository-enforced numeric threshold. Do not add tests only to raise a percentage.

## Artifacts and Final State

- Test harnesses may create `alloc/`, binaries, coverage files, and scoped environment data.
- Remove only task-created artifacts before completion. Never delete a pre-existing untracked file merely because it is under `alloc/`.
- `make worktree-test` owns cleanup for its isolated environment. Verify final cleanup with `git status --short`.

## Validation

For test-source changes, run the smallest focused command first, then the relevant broader Makefile target. For any code change, the final gate is:

```bash
make worktree-test
git diff --check
git status --short
```

For rule-only changes, use the rule validator plus diff/status inspection; do not run the test suite solely for Markdown.
