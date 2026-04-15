# CLAUDE.md — goth-scaffold

Go backend scaffold. Replace this file with project-specific instructions when you clone the scaffold.

## Common Commands

```bash
# ── First-time / after env-clean ──────────────────────────────────────────
make local-db-seed       # setup config + start services + wait for MySQL + migrate + seed

# ── Daily dev ─────────────────────────────────────────────────────────────
make local-env-start     # Start Docker services (MySQL / Redis / Cassandra / Asynqmon)
make local-env-stop      # Stop Docker services
make build               # Compile binary for current OS
make local-run           # Build and start API on :8080 (default mode)

# ── Testing ───────────────────────────────────────────────────────────────
make local-test          # Unit tests only (no services required)
make local-test-e2e      # E2E tests (services must be up; starts app server automatically)

# ── Maintenance ───────────────────────────────────────────────────────────
make local-db-reseed     # Drop + recreate MySQL DB, re-run migrations and seeds (services must be up)
make local-env-clean     # Destroy local data volumes and config (destructive, no prompt)
make scaffold            # Create new project from this scaffold (interactive)
```

Full command list: `make help`

## Run Modes

The `-m` flag controls which daemons activate:

| Mode           | Purpose                                |
|----------------|----------------------------------------|
| `default`      | API + worker + all daemons (local dev) |
| `api`          | HTTP server only                       |
| `worker`       | Asynq worker only                      |
| `db_migration` | Run migrations and exit                |
| `db_seed`      | Run seed and exit                      |

## Architecture

```
main.go → app.Initialize() → daemons.LoadActiveService()
```

Daemons in `app/daemons/` execute in **filename-prefix order**:

- `01–07`: environment, logger, profiler, MySQL, Redis, HTTP session
- `10–22`: Cassandra, worker setup, scheduler
- `100`: scheduled health-check (self-monitoring)
- `91–96`: DB migration, DB seed, API start, worker start
- `9999`: graceful shutdown

Request pipeline:

```
ghttp.Request → route.go → acceptances/ → endpoints/ → services/ → repositories/ → models/
```

## Components (`app/components/`)

Shared, stateless utility packages — import freely anywhere:

| Package       | Purpose                                                    |
|---------------|------------------------------------------------------------|
| `crypto`      | ID encryption/decryption (`EncryptKeyId` / `DecryptKeyId`) |
| `googleauth`  | Google OAuth token verification                            |
| `httpclient`  | Thin HTTP client wrapper with timeout/retry                |
| `queryfilter` | RSQL-style query string → GORM scope parser                |
| `recaptcha`   | Google reCAPTCHA v2/v3 verification                        |
| `slack`       | Slack incoming webhook sender                              |

## Key Conventions

- **Private struct fields/methods**: `_PascalCase` (e.g. `_Repo`, `_buildQuery`)
- **Repository methods must NOT return errors** — return `nil`/empty on failure, log via `kklogger.ErrorJ`
- **Logger format**: `package:Struct.Method#section!action_tag` (English only)
- **Never return `{"success": true}`** — HTTP 200 + `nil` body is the success signal
- **ID encryption**: use `crypto.EncryptKeyId` / `crypto.DecryptKeyId` — never roll your own
- **Blank line after `}`** before the next statement

## Directory CLAUDE.md Files

Each major directory has its own focused rules:

- `app/models/CLAUDE.md` — ID types, GORM tags, lazy associations
- `app/repositories/CLAUDE.md` — Repository pattern, FirstWhere/FindWhere
- `app/services/CLAUDE.md` — Service layer, Tx methods, logging
- `app/worker/CLAUDE.md` — Task/payload contract, asynqmon
- `app/connector/CLAUDE.md` — Database/Redis/Cassandra connectors, Enabled() pattern
- `app/database/migrate/CLAUDE.md` — Migration naming, up/down pairs
- `app/database/seed/CLAUDE.md` — Seed interface, order conventions, idempotency
- `app/handlers/CLAUDE.md` — Handler contract, route inheritance
- `app/handlers/endpoints/CLAUDE.md` — Dispatch order, file sections, inherited helpers
- `tests/CLAUDE.md` — Mandatory flags, output policy
- `tests/e2e/CLAUDE.md` — E2E boilerplate, prohibited patterns, edge-case checklist

## Git Workflow

- Branch: `feature/*`, `fix/*`, `chore/*, db/*, image/*`
- Commit messages: `feat(scope): 中文說明` (conventional commits)
- Target `staging` for MRs; `staging` → `main` for releases

## Module Name

`github.com/yetiz-org/goth-scaffold` — replace with your own via `make scaffold`.

## Evaluate Directory

`evaluate/` contains Docker Compose–based integration tests for CI pipelines.
Run via `make evaluate` — spins up full infrastructure, runs the app, verifies health.
Do not edit files in `evaluate/` during normal feature development.
