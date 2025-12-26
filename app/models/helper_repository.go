package models

import (
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrUniqueCreateNotApplied = errors.Errorf("unique create is not applied")
var ErrModelMetadataNotFound = errors.Errorf("entity model metadata is not found")

type SavableRepository[T Model] interface {
	Save(entity T) error
}

type SavableTxRepository[T Model] interface {
	SaveTx(tx *gorm.DB, entity T) error
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
	QueryBuilder() *QueryBuilder[T]
	SaveQuery(entity T) (stmt string, args []any)
	DeleteQuery(entity T) (stmt string, args []any)
}

type DatabaseRepository[T Model] interface {
	Repository[T]
	SavableTxRepository[T]
	DeletableTxRepository[T]
	UpsertableRepository[T]
	UpsertableTxRepository[T]
	FirstOrCreatableRepository[T]
	FirstOrCreatableTxRepository[T]
	DB() *gorm.DB
	DefaultLimit() int
}
