package repositories

import (
	"fmt"
	"strings"

	kklogger "github.com/yetiz-org/goth-kklogger"
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

func (q *_GormQuerier) FindHasManyComposite(dest any, conditions map[string]any, extra ...models.LazyCondition) error {
	db := _ApplyEqAnd(q._DbFn(), conditions)
	db = _ApplyLazyConds(db, extra)
	return db.Find(dest).Error
}

func (q *_GormQuerier) FindHasOneComposite(dest any, conditions map[string]any, extra ...models.LazyCondition) error {
	db := _ApplyEqAnd(q._DbFn(), conditions)
	db = _ApplyLazyConds(db, extra)
	return db.First(dest).Error
}

func (q *_GormQuerier) FindHasManyInComposite(dest any, fkColumns []string, compositeKeys [][]any, extra ...models.LazyCondition) error {
	if len(compositeKeys) == 0 {
		return nil
	}

	cols := "(" + strings.Join(fkColumns, ", ") + ")"
	db := q._DbFn().Where(cols+" IN ?", compositeKeys)
	db = _ApplyLazyConds(db, extra)
	return db.Find(dest).Error
}

// _ApplyEqAnd appends `col = ?` for each entry in conditions.
func _ApplyEqAnd(db *gorm.DB, conditions map[string]any) *gorm.DB {
	for col, val := range conditions {
		db = db.Where(col+" = ?", val)
	}

	return db
}

// _ApplyLazyConds appends extra scope conditions (any operator) as AND clauses.
// Unsupported operators are skipped with an error log so a tag typo does not silently
// widen the result set; the load still proceeds with the FK match only.
func _ApplyLazyConds(db *gorm.DB, conds []models.LazyCondition) *gorm.DB {
	for _, c := range conds {
		clause, args, ok := _RenderLazyCond(c)
		if !ok {
			kklogger.ErrorJ("repositories:_GormQuerier._ApplyLazyConds#unsupported_op",
				fmt.Sprintf("column=%q op=%q", c.Column, c.Op))
			continue
		}

		db = db.Where(clause, args...)
	}

	return db
}

// _RenderLazyCond renders a LazyCondition into a parameterized WHERE clause. The column is
// quoted via the active dialect (_QuoteSQLIdentifier) so the clause is valid on MySQL and
// PostgreSQL alike. ok is false for an unrecognized operator.
func _RenderLazyCond(c models.LazyCondition) (clause string, args []any, ok bool) {
	col := _QuoteSQLIdentifier(c.Column)
	switch c.Op {
	case models.LazyOpEq:
		return col + " = ?", []any{c.Value}, true
	case models.LazyOpNe:
		return col + " != ?", []any{c.Value}, true
	case models.LazyOpGt:
		return col + " > ?", []any{c.Value}, true
	case models.LazyOpGte:
		return col + " >= ?", []any{c.Value}, true
	case models.LazyOpLt:
		return col + " < ?", []any{c.Value}, true
	case models.LazyOpLte:
		return col + " <= ?", []any{c.Value}, true
	case models.LazyOpLike:
		return col + " LIKE ?", []any{c.Value}, true
	case models.LazyOpPrefix:
		return col + " LIKE ?", []any{fmt.Sprintf("%v%%", c.Value)}, true
	case models.LazyOpSuffix:
		return col + " LIKE ?", []any{fmt.Sprintf("%%%v", c.Value)}, true
	case models.LazyOpEmpty:
		return "COALESCE(" + col + ", '') = ''", nil, true
	case models.LazyOpNotEmpty:
		return "COALESCE(" + col + ", '') != ''", nil, true
	default:
		return "", nil, false
	}
}

func init() {
	q := &_GormQuerier{_DbFn: database.Reader}
	models.RegisterLazyQuerier(q)
}
