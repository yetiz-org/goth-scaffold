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
	_DbFn func() *gorm.DB
}

func (q *_GormQuerier) FindById(dest any, id any) error {
	return q._DbFn().First(dest, id).Error
}

func (q *_GormQuerier) FindHasMany(dest any, fkColumn string, fkValue any) error {
	return q._DbFn().Where(fkColumn+" = ?", fkValue).Find(dest).Error
}

func (q *_GormQuerier) FindHasOne(dest any, fkColumn string, fkValue any) error {
	return q._DbFn().Where(fkColumn+" = ?", fkValue).First(dest).Error
}

func (q *_GormQuerier) FindByIds(dest any, ids any) error {
	return q._DbFn().Where("id IN ?", ids).Find(dest).Error
}

func (q *_GormQuerier) FindHasManyIn(dest any, fkColumn string, fkValues any) error {
	return q._DbFn().Where(fkColumn+" IN ?", fkValues).Find(dest).Error
}

func (q *_GormQuerier) FindHasManyComposite(dest any, conditions map[string]any) error {
	db := q._DbFn()
	for col, val := range conditions {
		db = db.Where(col+" = ?", val)
	}

	return db.Find(dest).Error
}

func (q *_GormQuerier) FindHasOneComposite(dest any, conditions map[string]any) error {
	db := q._DbFn()
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

	return q._DbFn().Where(cols+" IN ?", compositeKeys).Find(dest).Error
}

func init() {
	q := &_GormQuerier{_DbFn: database.Reader}
	models.RegisterLazyQuerier(q)
}
