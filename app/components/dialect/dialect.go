package dialect

import (
	"database/sql"
	"sync"

	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"gorm.io/gorm"
)

type Name string

const (
	NameMySQL    Name = "mysql"
	NamePostgres Name = "postgres"
)

// Dialect abstracts all SQL dialect-specific behaviour so callers can swap
// between MySQL and PostgreSQL by changing the JSON secret `adapter` value
// alone. Registered implementations live in this package.
type Dialect interface {
	Name() Name
	QuoteIdent(ident string) string
	LikeOperator() string
	IsDuplicateKeyErr(err error) bool
	IsLockNoWaitErr(err error) bool
	IsRetryableErr(err error) bool
	MigrationDriver(sqlDB *sql.DB) (driver migratedb.Driver, dialectName string, err error)
	MigrationSchemaDir() string
}

var (
	registry = map[Name]Dialect{}
	regMu    sync.RWMutex

	cachedDialect Dialect
	cacheMu       sync.RWMutex
)

func Register(d Dialect) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[d.Name()] = d
}

func lookup(name Name) (Dialect, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	d, ok := registry[name]
	return d, ok
}

// Current returns the dialect matching the active GORM connection. The
// adapter name is derived from `database.Instance().Reader().Dialector.Name()`,
// which in turn comes from the JSON secret file (`adapter` field).
//
// Falls back to MySQL when no database is initialised — keeps migrations and
// seed daemons usable without a running DB.
func Current() Dialect {
	cacheMu.RLock()
	d := cachedDialect
	cacheMu.RUnlock()
	if d != nil {
		return d
	}

	resolved := resolveFromGORM()
	cacheMu.Lock()
	cachedDialect = resolved
	cacheMu.Unlock()
	return resolved
}

// ResetCache clears the cached dialect. Tests must call this when swapping
// the underlying GORM connection — production callers never need it.
func ResetCache() {
	cacheMu.Lock()
	cachedDialect = nil
	cacheMu.Unlock()
}

func resolveFromGORM() Dialect {
	if !database.Enabled() {
		if d, ok := lookup(NameMySQL); ok {
			return d
		}
	}

	reader := database.Reader()
	if reader == nil {
		if d, ok := lookup(NameMySQL); ok {
			return d
		}
	}

	return fromGORM(reader)
}

func fromGORM(db *gorm.DB) Dialect {
	if db == nil || db.Dialector == nil {
		d, _ := lookup(NameMySQL)
		return d
	}

	switch Name(db.Dialector.Name()) {
	case NamePostgres:
		if d, ok := lookup(NamePostgres); ok {
			return d
		}
	case NameMySQL:
		if d, ok := lookup(NameMySQL); ok {
			return d
		}
	}

	d, _ := lookup(NameMySQL)
	return d
}
