package dialect

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/go-sql-driver/mysql"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
)

type mysqlDialect struct{}

func (mysqlDialect) Name() Name { return NameMySQL }

func (mysqlDialect) QuoteIdent(ident string) string {
	return quoteIdent(ident, "`", "`")
}

func (mysqlDialect) LikeOperator() string { return "LIKE" }

func (mysqlDialect) IsDuplicateKeyErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (mysqlDialect) IsLockNoWaitErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 3572
}

func (mysqlDialect) IsRetryableErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1213, 1205: // Deadlock, Lock wait timeout
			return true
		}
	}

	msg := err.Error()
	return strings.Contains(msg, "Deadlock") ||
		strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "Lock wait timeout")
}

func (mysqlDialect) MigrationDriver(sqlDB *sql.DB) (migratedb.Driver, string, error) {
	driver, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	if err != nil {
		return nil, "", err
	}
	return driver, string(NameMySQL), nil
}

func (mysqlDialect) MigrationSchemaDir() string { return string(NameMySQL) }

// quoteIdent applies open/close quote pairs to each dot-separated segment.
// Empty identifier, already-quoted identifier, or identifier with special
// characters is returned unchanged.
func quoteIdent(ident, open, close string) string {
	if ident == "" {
		return ident
	}

	if strings.Contains(ident, open) || strings.Contains(ident, close) {
		return ident
	}

	for _, r := range ident {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.') {
			return ident
		}
	}

	parts := strings.Split(ident, ".")
	for i, p := range parts {
		if p == "" {
			return ident
		}
		parts[i] = fmt.Sprintf("%s%s%s", open, p, close)
	}

	return strings.Join(parts, ".")
}

func init() { Register(mysqlDialect{}) }
