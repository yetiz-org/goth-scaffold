# app/connector

## Available Connectors

| Package               | Daemon                  | Config key                | Backend           |
|-----------------------|-------------------------|---------------------------|-------------------|
| `connector/database`  | `05_setup_database.go`  | `DataStore.DatabaseName`  | MySQL (GORM)      |
| `connector/redis`     | `06_setup_redis.go`     | `DataStore.RedisName`     | Redis             |
| `connector/keyspaces` | `10_setup_keyspaces.go` | `DataStore.CassandraName` | Cassandra (gocql) |

## Pattern

Each connector exposes a package-level `Instance()` that returns the singleton, plus named shortcuts:

```go
func Instance() *datastore.Database { _Init(); return _db }
func Reader() *gorm.DB              { return Instance().Reader().DB() }
func Writer() *gorm.DB              { return Instance().Writer().DB() }
func Enabled() bool                 { return conf.Config().DataStore.DatabaseName != "" }
func HealthCheck() error            { ... }
```

Cassandra (`connector/keyspaces`) follows the same shape but returns `datastore.CassandraOperator`:

```go
func Writer() datastore.CassandraOperator { return Instance().Writer() }
func Reader() datastore.CassandraOperator { return Instance().Reader() }
```

- Initialisation is guarded by `sync.Once`
- **Always call `Enabled()` before first use** — each daemon (`05`, `06`, `10`) guards its connector
- All errors are logged via `kklogger.ErrorJ`; functions never panic

## Adding a New Connector

1. Create `app/connector/<name>/<name>.go`
2. Follow the `Instance / Reader / Writer / Enabled / HealthCheck` pattern
3. Wire up in `app/daemons/0N_setup_<name>.go`
4. Call `Enabled()` as the first guard in the daemon — skip init silently when config is absent:

```go
func SetupFoo(ctx context.Context) error {
	if !foo.Enabled() {
		return nil
	}
	// init...
	return nil
}
```

## Redis Key Convention

```go
redis.Key("CATEGORY", "sub-key")
// → APP_NAME:ENVIRONMENT:CATEGORY:sub-key
```

Always use `redis.Key()` — never construct keys manually.
