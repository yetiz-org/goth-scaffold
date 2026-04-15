package models

import (
	"math"
	"strings"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrUniqueCreateNotApplied = errors.Errorf("unique create is not applied")
var ErrModelMetadataNotFound = errors.Errorf("entity model metadata is not found")

type SavableRepository[T Model] interface {
	Save(entity T) error
}

type SaveRetryRepository[T Model] interface {
	SaveRetry(entity T) error
}

type SavableTxRepository[T Model] interface {
	SaveTx(tx *gorm.DB, entity T) error
}

type SaveRetryTxRepository[T Model] interface {
	SaveRetryTx(tx *gorm.DB, entity T) error
}

type DeletableRepository[T Model] interface {
	Delete(entity T) error
}

type DeletableTxRepository[T Model] interface {
	DeleteTx(tx *gorm.DB, entity T) error
}

type UpsertableRepository[T Model] interface {
	Upsert(entity T, conditions map[string]any) error
}

type UpsertableTxRepository[T Model] interface {
	UpsertTx(tx *gorm.DB, entity T, conditions map[string]any) error
}

type FirstOrCreatableRepository[T Model] interface {
	FirstOrCreate(entity T, conditions map[string]any) (bool, error)
}

type FirstOrCreatableTxRepository[T Model] interface {
	FirstOrCreateTx(tx *gorm.DB, entity T, conditions map[string]any) (bool, error)
}

type FetchableRepository[T Model] interface {
	Fetch(id any, opts ...DatabaseQueryOption[T]) (model T, err error)
}

type FindableRepository[T Model] interface {
	Find(opts ...DatabaseQueryOption[T]) ([]T, error)
	First(opts ...DatabaseQueryOption[T]) (T, error)
	FindWhere(build func(*gorm.DB) *gorm.DB, opts ...DatabaseQueryOption[T]) ([]T, error)
	FirstWhere(build func(*gorm.DB) *gorm.DB, opts ...DatabaseQueryOption[T]) (T, error)
}

type GettableRepository[K any, T Model] interface {
	Get(id K, opts ...DatabaseQueryOption[T]) T
}

type TryUnique[T Model] interface {
	UniqueCreate(entity T) error
}

type TryResult struct {
	LastApplied bool
	LastResult  map[string]any
	Error       error
}

type Repository[T Model] interface {
	TableName() string
	SavableRepository[T]
	DeletableRepository[T]
}

type CassandraRepository[T Model] interface {
	Repository[T]
	TryUnique[T]
	Session() *gocql.Session
	QueryBuilder() *CassandraQueryBuilder[T]
	SaveQuery(entity T) (stmt string, args []any)
	DeleteQuery(entity T) (stmt string, args []any)
}

type DatabaseRepository[K any, T Model] interface {
	Repository[T]
	SavableTxRepository[T]
	SaveRetryRepository[T]
	SaveRetryTxRepository[T]
	DeletableTxRepository[T]
	UpsertableRepository[T]
	UpsertableTxRepository[T]
	FirstOrCreatableRepository[T]
	FirstOrCreatableTxRepository[T]
	FetchableRepository[T]
	FindableRepository[T]
	GettableRepository[K, T]
	DB() *gorm.DB
	DefaultLimit() int
}

// DatabaseQueryOption wraps a GORM scope, an optional eager-load function, and an
// optional in-memory filter/sort function into a single composable value.
// T must be a pointer model type (e.g. *SiteSetting).
// _GormFn runs before the query; _LoadFn runs after (opt-in, only when provided);
// _FilterFn runs after the query and returns a filtered/sorted copy of the slice.
type DatabaseQueryOption[T any] struct {
	_GormFn     func(*gorm.DB) *gorm.DB
	_LoadFn     func([]T)
	_FilterFn   func([]T) []T
	_SelectCols []string
}

// SelectCols returns the column list specified by this option (with the primary key
// automatically included). Returns nil when SelectOpt was not used.
func (o DatabaseQueryOption[T]) SelectCols() []string {
	return o._SelectCols
}

// ApplyGorm applies the GORM scope. No-op when _GormFn is nil.
func (o DatabaseQueryOption[T]) ApplyGorm(db *gorm.DB) *gorm.DB {
	if o._GormFn != nil {
		return o._GormFn(db)
	}

	return db
}

// ApplyEager runs the eager-load function against items. No-op when _LoadFn is nil.
func (o DatabaseQueryOption[T]) ApplyEager(items []T) {
	if o._LoadFn != nil {
		o._LoadFn(items)
	}
}

// ApplyFilter runs the in-memory filter/sort function. Returns items unchanged when _FilterFn is nil.
func (o DatabaseQueryOption[T]) ApplyFilter(items []T) []T {
	if o._FilterFn != nil {
		return o._FilterFn(items)
	}

	return items
}

// GormOpt wraps a func(*gorm.DB)*gorm.DB into a DatabaseQueryOption.
func GormOpt[T any](fn func(*gorm.DB) *gorm.DB) DatabaseQueryOption[T] {
	return DatabaseQueryOption[T]{_GormFn: fn}
}

// QueryOpt creates a DatabaseQueryOption with both a GORM scope and an in-memory filter.
// Used by the queryfilter component; either gormFn or filterFn may be nil.
func QueryOpt[T any](gormFn func(*gorm.DB) *gorm.DB, filterFn func([]T) []T) DatabaseQueryOption[T] {
	return DatabaseQueryOption[T]{_GormFn: gormFn, _FilterFn: filterFn}
}

// PaginationOpt wraps offset and limit into a DatabaseQueryOption that applies LIMIT/OFFSET
// at the DB layer. MySQL requires LIMIT when OFFSET is used; when offset > 0 and limit <= 0,
// a large sentinel LIMIT (math.MaxInt32) is applied so the query remains valid.
func PaginationOpt[T any](offset, limit int) DatabaseQueryOption[T] {
	return DatabaseQueryOption[T]{_GormFn: func(db *gorm.DB) *gorm.DB {
		if offset > 0 {
			if limit <= 0 {
				db = db.Limit(math.MaxInt32)
			}

			db = db.Offset(offset)
		}

		if limit > 0 {
			db = db.Limit(limit)
		}

		return db
	}}
}

// LimitOpt wraps limit into a DatabaseQueryOption that applies LIMIT at the DB layer (no OFFSET).
func LimitOpt[T any](limit int) DatabaseQueryOption[T] {
	return DatabaseQueryOption[T]{_GormFn: func(db *gorm.DB) *gorm.DB {
		if limit > 0 {
			db = db.Limit(limit)
		}

		return db
	}}
}

// EagerAll returns a DatabaseQueryOption that auto-scans all lazy associations and
// batch-preloads them. Preloading is opt-in: it only runs when this option is passed.
func EagerAll[T any]() DatabaseQueryOption[T] {
	return DatabaseQueryOption[T]{_LoadFn: _autoEagerLoad[T]}
}

// SelectOpt restricts the query to specific columns. The primary key (id) is always included.
func SelectOpt[T any](columns ...string) DatabaseQueryOption[T] {
	cols := _EnsurePrimaryKey(columns)
	return DatabaseQueryOption[T]{
		_GormFn:     func(db *gorm.DB) *gorm.DB { return db.Select(cols) },
		_SelectCols: cols,
	}
}

// _EnsurePrimaryKey returns columns with "id" prepended if not already present.
func _EnsurePrimaryKey(columns []string) []string {
	for _, c := range columns {
		if strings.EqualFold(c, "id") {
			result := make([]string, len(columns))
			copy(result, columns)
			return result
		}
	}

	result := make([]string, 0, len(columns)+1)
	result = append(result, "id")
	result = append(result, columns...)
	return result
}
