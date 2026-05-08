# evaluate

This directory provides local infrastructure for development and E2E tests.

It can start:
- MySQL or PostgreSQL, selected by `DB_ADAPTER`
- Redis
- Cassandra
- Asynqmon

## Local Scope

Local commands use:
- secrets: `evaluate/env/local`
- runtime data and generated config: `evaluate/_run/local`
- default host ports: `3306`, `5432`, `6379`, `9042`, `8081`, app `8080`

```bash
make local-env-setup
make local-env-start
make local-env-status
make local-env-stop
make local-env-clean
```

## Worktree Scope

Worktree commands use:
- secrets: `evaluate/env/worktree/<safe-id>`
- runtime data and generated config: `evaluate/_run/worktree/<safe-id>`
- isolated compose project and container names
- host ports persisted in `evaluate/_run/worktree/<safe-id>/ports.env`

```bash
make worktree-env-setup WORKTREE_ID=my-branch
make worktree-env-start WORKTREE_ID=my-branch
make worktree-env-status WORKTREE_ID=my-branch
make worktree-env-stop WORKTREE_ID=my-branch
make worktree-env-clean WORKTREE_ID=my-branch
```

When `WORKTREE_ID` is omitted, the Makefile derives it from the current directory name.
The generated `<safe-id>` is a sanitized id plus a checksum of the raw `WORKTREE_ID`,
so similar names such as `feature/foo` and `feature-foo` remain isolated.
The port allocator starts from a checksum-derived block, then skips ports already
reserved by other worktree scopes or already listening on localhost.

## Tests

```bash
make local-test-e2e
make worktree-test-e2e WORKTREE_ID=my-branch
```

The Makefile passes the matching scoped config to the E2E harness through `SCAFFOLD_E2E_CONFIG`.
