# goth-scaffold Rules

## Purpose

goth-scaffold is a reusable Go backend scaffold. Keep examples, defaults, docs, and rules generic. Do not add product, customer, vendor, dataset, or business-workflow names to scaffold-owned examples.

## Priorities

1. Preserve public behavior, compatibility, data safety, and lifecycle correctness.
2. Keep ownership clear across handlers, services, repositories, models, workers, and components.
3. Prefer the smallest complete change backed by current source and tests.
4. Keep generated configuration and test environments isolated by Makefile scope.

## Working Boundaries

- Read the complete target file, its applicable subtree rules, callers, non-call references, and existing tests before changing behavior.
- Do not edit `.agents/`, `.claude/`, rule-file topology, generated config, or `evaluate/` unless the task explicitly targets that path.
- Do not create or switch Git worktrees unless the user requests it.
- Do not hand-create or commit scoped runtime config. Use Makefile targets, which write under `evaluate/env/<scope>` and `evaluate/_run/<scope>`.
- For prose-only changes, run structural/document checks rather than Go tests. For Go source or dependency changes, use the code gates below.

## Common Commands

| Goal | Command | Notes |
| --- | --- | --- |
| First local bootstrap | `make local-db-seed` | Creates scoped config, starts services, migrates, and seeds. |
| Start or stop services | `make local-env-start` / `make local-env-stop` | Uses `DB_ADAPTER` and the active scope. |
| Inspect services | `make local-env-status` | Shows the active Compose scope. |
| Build or run | `make build` / `make local-run` | `local-run` starts the default mode after environment setup. |
| Generate OpenAPI | `make docs-build` | Runs `cmd/goaispec` and writes `docs/openapi/openapi.yaml`. |
| Unit tests | `make local-test` | Runs `go test -v -count=1 ./tests/units/...`. |
| Isolated E2E | `make worktree-test-e2e` | Creates worktree-scoped ports, config, and services. |
| Full isolated gate | `make worktree-test` | Cleans, seeds, runs race-enabled non-E2E tests and E2E, then cleans its scope. |
| Reset local data | `make local-db-reseed` | Destructive within the active local scope. |
| Destroy local scope | `make local-env-clean` | Removes scoped containers, volumes, config, and generated data. |
| List all targets | `make help` | Treat the current Makefile as command truth. |

Do not pipe, truncate, or discard test output. Preserve complete failure output for diagnosis.

## Runtime and Architecture

`main.go` calls `app.Initialize()`, which parses `-c` and `-m`, loads config, and builds the active daemon service. `cmd/goaispec` is the documentation-only command; it builds the route tree with handler registration disabled and must not start connectors or daemons.

Run modes:

| Mode | Active role |
| --- | --- |
| `default` | HTTP API, database migration, schedulers, and supporting daemons; no worker setup or start |
| `api` | HTTP API and worker setup; no database migration or worker start |
| `worker` | Asynq worker and supporting daemons; no HTTP API |
| `db_migration` | Run migrations and exit |
| `db_seed` | Run seeds and exit |

Daemon order is declared in `app/daemons/daemon_preset*.go`; numeric filenames mirror that order but are not the source of registration. Register database consumers after the database daemon, network entry points after their dependencies, and shutdown last. Verify every mode that should own the new daemon.

Request flow:

```text
ghttp.Request -> route.go -> acceptances/ -> endpoints/ -> services/ -> repositories/ -> models/
```

Layer ownership:

- handlers own transport, parsing, acceptance chains, and response shape;
- services own application workflows and transaction coordination;
- repositories own persistence access;
- models own persistence and association shape;
- workers own asynchronous task execution;
- components own reusable stateless infrastructure.

Reuse an existing owner before adding a helper, package, wrapper, or duplicate policy.

## Project Conventions

- Private struct fields and methods use `_PascalCase`.
- Logger keys use `package:Struct.Method#section!action`; `action` requires `section`, and keys/messages are English.
- HTTP success with no payload is status `200` and a nil body; never add `{"success": true}`.
- JSON response bodies use a top-level object. Empty bodies, redirects, and non-JSON streams are exempt.
- Use `app/components/crypto` for ID encryption and decryption.
- Keep a blank line after a closing `}` before the next statement where project style requires it.
- Follow the deeper rule file for repository, model, handler, worker, connector, migration, seed, service, test, or docs details.

## Go Modules

When updating dependencies:

1. Enumerate the exact current requirements with `go mod edit -json`.
2. Query updates with `go list -m -u -json <module...>`; do not infer versions from another repository.
3. Review release/module diffs for imported packages and note behavior changes.
4. Apply only authorized modules, run `go mod tidy`, and inspect both `go.mod` and `go.sum`.
5. Run the build, focused tests, and full worktree gate. Stop on a required Go-version or public-contract change that was not authorized.

The `go` directive controls module language semantics. The selected `go version` controls the compiler and standard library; record both for version-sensitive changes.

## Directory Rules

| Scope | Rule |
| --- | --- |
| `app/models/` | `app/models/CLAUDE.md` |
| `app/repositories/` | `app/repositories/CLAUDE.md` |
| `app/services/` | `app/services/CLAUDE.md` |
| `app/worker/` | `app/worker/CLAUDE.md` |
| `app/connector/` | `app/connector/CLAUDE.md` |
| `app/database/migrate/` | `app/database/migrate/CLAUDE.md` |
| `app/database/seed/` | `app/database/seed/CLAUDE.md` |
| `app/handlers/` | `app/handlers/CLAUDE.md` |
| `app/handlers/endpoints/` | `app/handlers/endpoints/CLAUDE.md` |
| `tests/` | `tests/CLAUDE.md` |
| `tests/units/` | `tests/units/CLAUDE.md` |
| `tests/e2e/` | `tests/e2e/CLAUDE.md` |
| `docs/` | `docs/CLAUDE.md` |

## Validation

| Change | Required checks |
| --- | --- |
| Rule or docs only | Applicable rule/doc validator, `git diff --check`, final diff and status inspection |
| Go dependency metadata | `go mod tidy`, `make build`, `make local-test`, `make worktree-test` |
| Go source | Formatter on changed files, focused tests, `make build`, `make worktree-test` |
| YAML | Parse every changed YAML file with Ruby or Python before broader checks |

Before completion, inspect the final tracked and untracked state. Remove only task-created residue; preserve pre-existing files and user work.
