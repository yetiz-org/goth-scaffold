# E2E Test Rules

## Purpose

Tests under `tests/e2e/` exercise the compiled application, HTTP stack, and configured connectors. The suite owns server startup through `TestMain`; the Makefile owns scoped config, ports, services, and environment cleanup.

## Running Tests

Use the isolated Makefile paths from the repository root:

```bash
make worktree-test-e2e
make worktree-test
```

`make local-test-e2e` is valid only after `make local-env-setup` and required services are available. A direct run must provide scoped artifacts explicitly:

```bash
SCAFFOLD_E2E_BINARY=./scaffold-amd64 SCAFFOLD_E2E_CONFIG=evaluate/_run/<scope>/config.yaml.local go test -v -count=1 -timeout=120s ./tests/e2e/...
```

Do not hand-create config files or reuse another worktree's `_run` path. The current Makefile does not implement a focused `E2E_RUN` variable.

## Test Shape

- Put one logical endpoint/resource group in each `{domain}_test.go` file.
- New or renamed functions use `Test<Domain>_<Operation>_<Scenario>` when those three parts apply. Keep names precise rather than padding a simple liveness test.
- Every `Test...` function calls `t.Parallel()` as its first statement. `TestMain` is not a test function and is exempt.
- After optional-prerequisite checks, create one `testutils.NewTestContext(t)` per test. Do not share clients or mutable state between tests.
- Close every response body. Use `testutils.ReadJSONBody` when decoding JSON.
- When creating or substantially editing a test file, add the synchronized file-level block comment required by `tests/CLAUDE.md`.

## Scenario Selection

For each changed endpoint, derive applicable cases from the actual route, acceptance chain, parsing, validation, and response code:

| Category | Examples |
| --- | --- |
| Success | Valid request returns the documented status and response fields. |
| Authentication/authorization | Missing, invalid, or insufficient credentials are rejected when the route requires them. |
| Input validation | Empty, malformed, oversized, or unsupported values are rejected. |
| Boundary | Minimum, maximum, empty collection, or pagination edge that the contract defines. |
| Missing/conflict | Unknown resource, duplicate operation, or invalid state transition when applicable. |
| Method | Unsupported HTTP verb returns the served method error when the router exposes one. |

Do not add irrelevant matrix cases. State why an omitted category does not apply when it is not obvious.

## Assertions

- Verify response fields and side effects, not only status, whenever the endpoint has a body or state contract.
- A status-only assertion is acceptable for a pure liveness/connectivity contract.
- Use fatal failures for prerequisites that make later checks unsafe; use non-fatal checks only when later assertions remain meaningful.
- Include expected and actual values in failure messages.
- Reject ranges so broad that broken behavior still passes.

## Isolation and Optional Services

- Generate unique test data. Do not hardcode reusable resource identifiers or share state across tests.
- `t.Skip` is allowed only when an explicitly optional connector/config is disabled or not initialized in a non-CI environment. The message must name the missing prerequisite and how to enable it.
- Once a required service is available, health or behavior failures are test failures, not skip conditions.
- CI and `make worktree-test` must not silently pass by skipping required setup.
- Do not use `time.Sleep` in test cases. Poll an observable condition with a deadline; reuse `testutils.WaitForServer` for server readiness.

## Test Utilities and Lifecycle

| API | Purpose |
| --- | --- |
| `testutils.NewTestContext(t)` | Per-test base URL and HTTP client |
| `tc.DoGet(path)` | GET relative to the scoped base URL |
| `tc.DoPostJSON(path, body)` | JSON POST relative to the scoped base URL |
| `tc.NewClient()` | Client that does not follow redirects |
| `testutils.ReadJSONBody(t, resp, &dest)` | Read, close, and decode a JSON body |
| `testutils.GetTestConfig()` | Resolve CI/local scoped config and base URL |
| `testutils.WaitForServer(url, timeout)` | Bounded readiness polling |

`TestMain` starts and stops the server and cleans its temporary binary/config. Do not add another package-wide lifecycle owner.

## Artifacts and Completion

- Server logs may appear under repository-root `alloc/logs/`. Preserve logs that predate the task.
- Remove only binaries, logs, configs, coverage files, or directories created by the current task and not owned by Makefile cleanup.
- After E2E or full-gate execution, inspect `git status --short` and report any residual artifact you cannot safely attribute.

For rule-only changes, run the rule validator plus diff/status checks instead of starting services.
