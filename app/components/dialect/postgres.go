package dialect

import (
	"database/sql"
	"errors"
	"strings"

	migratedb "github.com/golang-migrate/migrate/v4/database"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDialect struct{}

func (postgresDialect) Name() Name { return NamePostgres }

// QuoteIdent wraps each dot-separated segment with double quotes to match
// PostgreSQL identifier syntax.
func (postgresDialect) QuoteIdent(ident string) string {
	return quoteIdent(ident, `"`, `"`)
}

// LikeOperator returns ILIKE so PostgreSQL behaves case-insensitively, matching
// MySQL's default utf8mb4_unicode_ci collation behaviour.
func (postgresDialect) LikeOperator() string { return "ILIKE" }

func (postgresDialect) IsDuplicateKeyErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	// lib/pq style surfaces the SQLSTATE in the message
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func (postgresDialect) IsLockNoWaitErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "55P03" // lock_not_available
	}
	return strings.Contains(err.Error(), "SQLSTATE 55P03")
}

func (postgresDialect) IsRetryableErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001", "40P01": // serialization_failure, deadlock_detected
			return true
		}
	}

	msg := err.Error()
	return strings.Contains(msg, "serialization_failure") ||
		strings.Contains(msg, "deadlock_detected") ||
		strings.Contains(msg, "SQLSTATE 40001") ||
		strings.Contains(msg, "SQLSTATE 40P01")
}

func (postgresDialect) MigrationDriver(sqlDB *sql.DB) (migratedb.Driver, string, error) {
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return nil, "", err
	}
	return driver, string(NamePostgres), nil
}

func (postgresDialect) MigrationSchemaDir() string { return string(NamePostgres) }

func init() { Register(postgresDialect{}) }
