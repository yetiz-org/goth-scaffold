package repositories

import (
	"strings"

	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

// _GormQuerier implements models.LazyQuerier (and BatchQuerier, CompositeQuerier,
// CompositeBatchQuerier) using a GORM DB accessor function.
// This keeps GORM out of the models package and satisfies the dependency inversion principle.
type _GormQuerier struct {
	dbFn func() *gorm.DB
}

func (q *_GormQuerier) FindById(dest any, id any) error {
	return q.dbFn().First(dest, id).Error
}

func (q *_GormQuerier) FindHasMany(dest any, fkColumn string, fkValue any) error {
	return q.dbFn().Where(fkColumn+" = ?", fkValue).Find(dest).Error
}

func (q *_GormQuerier) FindHasOne(dest any, fkColumn string, fkValue any) error {
	return q.dbFn().Where(fkColumn+" = ?", fkValue).First(dest).Error
}

func (q *_GormQuerier) FindByIds(dest any, ids any) error {
	return q.dbFn().Where("id IN ?", ids).Find(dest).Error
}

func (q *_GormQuerier) FindHasManyIn(dest any, fkColumn string, fkValues any) error {
	return q.dbFn().Where(fkColumn+" IN ?", fkValues).Find(dest).Error
}

func (q *_GormQuerier) FindHasManyComposite(dest any, conditions map[string]any) error {
	db := q.dbFn()
	for col, val := range conditions {
		db = db.Where(col+" = ?", val)
	}

	return db.Find(dest).Error
}

func (q *_GormQuerier) FindHasOneComposite(dest any, conditions map[string]any) error {
	db := q.dbFn()
	for col, val := range conditions {
		db = db.Where(col+" = ?", val)
	}

	return db.First(dest).Error
}

func (q *_GormQuerier) FindHasManyInComposite(dest any, fkColumns []string, compositeKeys [][]any) error {
	if len(compositeKeys) == 0 {
		return nil
	}

	cols := "(" + strings.Join(fkColumns, ", ") + ")"

	return q.dbFn().Where(cols+" IN ?", compositeKeys).Find(dest).Error
}

func init() {
	if database.Enabled() {
		q := &_GormQuerier{dbFn: database.Reader}
		models.RegisterLazyQuerier(q)
	}
}
