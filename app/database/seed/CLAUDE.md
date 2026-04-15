# app/database/seed

## Interface

Every seed must implement:

```go
type Seed interface {
	Name() string   // unique identifier, snake_case
	Order() int     // execution priority — lower runs first
	Run() error     // returns non-nil only on fatal failure
}
```

## Registration

Use `init()` to auto-register at package import time:

```go
func init() { Register(&FooSeed{}) }
```

`Name()` must be **globally unique** across all seeds — it is used as the sort key within the same `Order()` bucket and
appears in all log output.

## RunAll Execution Model

`RunAll` (called by the `db_seed` daemon) follows this sequence:

1. Collect all registered seeds.
2. Sort: primary key `Order()` ascending, secondary key `Name()` ascending (alphabetical) — fully deterministic.
3. Execute each seed in order, logging start and duration.
4. If any seed returns a non-nil error → **halt immediately**, log the failure, and return the error. Subsequent seeds
   do not run.

```
Order 1 → Order 9 → Order 10 → Order 10 (alphabetical) → Order 20 …
```

## Order Conventions

| Range | Purpose                                                |
|-------|--------------------------------------------------------|
| 1–9   | Bootstrap (site keys, root users, credentials)         |
| 10–19 | Reference / lookup tables (roles, settings, constants) |
| 20–39 | Domain data (platforms, categories, defaults)          |
| 40+   | Test / sample data (dev/staging only)                  |

Seeds within the same order bucket are sorted alphabetically by name.

## Idempotency (critical)

Seeds run on every `db_seed` invocation. Always use **Upsert / FirstOrCreate** — never plain INSERT:

```go
func (s *FooSeed) Run() error {
	repo := repositories.NewFooRepository(database.Writer())
	for _, d := range defaults {
		item := &models.Foo{ID: models.FooId(d.id), Name: d.name}
		if err := repo.Upsert(item, map[string]any{"id": models.FooId(d.id)}); err != nil {
			kklogger.ErrorJ("seed:FooSeed.Run#upsert!db_error", err.Error())
			return err
		}
	}
	return nil
}
```

## Error Handling

- Return `nil` for non-fatal skips (e.g. already-seeded data with no diff).
- Return a non-nil `error` only when the seed cannot proceed and further seeds should halt.
- Log errors via `kklogger.ErrorJ("seed:FooSeed.Run#action!error_tag", ...)`.

## Logging

```go
kklogger.InfoJ("seed:FooSeed.Run#start!seed", "seeding foos")
kklogger.InfoJ("seed:FooSeed.Run#done!success", fmt.Sprintf("seeded %d foos", count))
```

Format: `seed:SeedName.Run#section!action` — English only.
