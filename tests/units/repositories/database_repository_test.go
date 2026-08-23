/*
Tests generic SQL repository batching and cancellable transaction retry behavior.
GORM dry-run sessions prove emitted MySQL/PostgreSQL query shape without opening a database.
*/

package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/yetiz-org/goth-scaffold/app/repositories"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type batchRecord struct {
	ID uint64 `gorm:"column:id;primaryKey"`
}

func (r *batchRecord) TableName() (name string) {
	return "batch_records"
}

type capturedQuery struct {
	SQL  string
	Vars []any
}

func newDryRunDB(t *testing.T, adapter string) (db *gorm.DB, captured *[]capturedQuery) {
	t.Helper()

	connection := &sql.DB{}
	var dialector gorm.Dialector
	switch adapter {
	case "mysql":
		dialector = mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true})
	case "postgres":
		dialector = postgres.New(postgres.Config{Conn: connection})
	default:
		t.Fatalf("unsupported adapter %q", adapter)
	}

	db, err := gorm.Open(dialector, &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	queries := make([]capturedQuery, 0, 1)
	if err := db.Callback().Query().After("gorm:query").Register("tests:capture_query", func(tx *gorm.DB) {
		queries = append(queries, capturedQuery{
			SQL:  tx.Statement.SQL.String(),
			Vars: append([]any(nil), tx.Statement.Vars...),
		})
	}); err != nil {
		t.Fatalf("register query callback error = %v", err)
	}

	return db, &queries
}

func TestDatabaseDefaultRepository_GetByIDs_UsesOneBoundedQueryAcrossDialects(t *testing.T) {
	for _, adapter := range []string{"mysql", "postgres"} {
		t.Run(adapter, func(t *testing.T) {
			db, captured := newDryRunDB(t, adapter)
			repository := repositories.NewDatabaseDefaultRepository[uint64, *batchRecord](db)

			if got := repository.GetByIDs(nil); len(got) != 0 {
				t.Fatalf("GetByIDs(nil) length = %d, want 0", len(got))
			}

			if len(*captured) != 0 {
				t.Fatalf("GetByIDs(nil) executed %d queries, want 0", len(*captured))
			}

			if got := repository.GetByIDs([]uint64{11, 22, 33}); len(got) != 0 {
				t.Fatalf("dry-run GetByIDs() length = %d, want 0", len(got))
			}

			if len(*captured) != 1 {
				t.Fatalf("GetByIDs() executed %d queries, want 1", len(*captured))
			}

			query := (*captured)[0]
			if !strings.Contains(query.SQL, "WHERE id IN") && !strings.Contains(query.SQL, "WHERE \"id\" IN") && !strings.Contains(query.SQL, "WHERE `id` IN") {
				t.Fatalf("GetByIDs() SQL = %q, want an id IN predicate", query.SQL)
			}

			wantVars := []any{uint64(11), uint64(22), uint64(33)}
			if !reflect.DeepEqual(query.Vars, wantVars) {
				t.Fatalf("GetByIDs() vars = %#v, want %#v", query.Vars, wantVars)
			}
		})
	}
}

func TestWithTransactionRetryContext_ReturnsCancellationWithoutAnotherAttempt(t *testing.T) {
	t.Chdir(t.TempDir())

	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	err := repositories.WithTransactionRetryContext(ctx, 3, func() (*gorm.DB, error) {
		attempts++
		cancel()
		return nil, &gomysql.MySQLError{Number: 1213, Message: "deadlock"}
	}, func(tx *gorm.DB) error {
		t.Fatal("transaction body ran after begin failed")
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithTransactionRetryContext() error = %v, want context.Canceled", err)
	}

	if attempts != 1 {
		t.Fatalf("begin attempts = %d, want 1", attempts)
	}
}

func TestWithTransactionRetryContext_RejectsNilContextBeforeBegin(t *testing.T) {
	attempts := 0
	err := repositories.WithTransactionRetryContext(nil, 3, func() (*gorm.DB, error) {
		attempts++
		return nil, nil
	}, func(tx *gorm.DB) error {
		t.Fatal("transaction body ran with a nil context")
		return nil
	})

	if err == nil || err.Error() != "transaction retry context required" {
		t.Fatalf("WithTransactionRetryContext(nil) error = %v, want context-required error", err)
	}

	if attempts != 0 {
		t.Fatalf("begin attempts = %d, want 0", attempts)
	}
}
