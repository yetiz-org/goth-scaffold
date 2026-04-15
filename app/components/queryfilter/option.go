package queryfilter

import (
	"fmt"
	"slices"
	"sort"
	"strconv"

	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

// ToOption converts a parsed filter AST and sort fields into a models.DatabaseQueryOption[T]
// using the provided schema for field resolution and validation.
//
// DB-safe conditions are applied via _GormFn (pre-query).
// In-memory conditions and sorts are applied via _FilterFn (post-query).
func ToOption[T any](node Node, sorts []SortField, schema Schema[T]) (models.DatabaseQueryOption[T], error) {
	if err := ValidateNode(node, schema); err != nil {
		return models.DatabaseQueryOption[T]{}, err
	}

	if err := ValidateSorts(sorts, schema); err != nil {
		return models.DatabaseQueryOption[T]{}, err
	}

	gormFn := buildGORMScope(node, sorts, schema)
	filterFn := buildInMemoryFn(node, sorts, schema)

	return models.QueryOpt(gormFn, filterFn), nil
}

// ValidateNode checks that all fields and operators in the AST are allowed by the schema.
func ValidateNode[T any](node Node, schema Schema[T]) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ComparisonNode:
		def, ok := schema[n.Field]
		if !ok {
			return fmt.Errorf("queryfilter: unknown field %q", n.Field)
		}

		if def.Column == "" && def.SQLExpr == nil && def.FilterFn == nil {
			return fmt.Errorf("queryfilter: field %q is not filterable", n.Field)
		}

		if len(def.AllowedOps) > 0 && !opsContain(def.AllowedOps, n.Op) {
			return fmt.Errorf("queryfilter: operator %q not allowed for field %q", n.Op, n.Field)
		}

		if def.Type == FieldTypeInt {
			if _, err := strconv.ParseInt(n.Value, 10, 64); err != nil {
				return fmt.Errorf("queryfilter: field %q expects integer value, got %q", n.Field, n.Value)
			}
		}

		if def.Type == FieldTypeBool {
			if _, err := strconv.ParseBool(n.Value); err != nil {
				return fmt.Errorf("queryfilter: field %q expects boolean value, got %q", n.Field, n.Value)
			}
		}

		if def.Type == FieldTypeFloat {
			if _, err := strconv.ParseFloat(n.Value, 64); err != nil {
				return fmt.Errorf("queryfilter: field %q expects decimal value, got %q", n.Field, n.Value)
			}
		}

		return nil

	case *LogicalNode:
		if err := ValidateNode(n.Left, schema); err != nil {
			return err
		}

		return ValidateNode(n.Right, schema)
	}

	return nil
}

// ValidateSorts checks that all sort fields are defined and sortable in the schema.
func ValidateSorts[T any](sorts []SortField, schema Schema[T]) error {
	for _, s := range sorts {
		def, ok := schema[s.Field]
		if !ok {
			return fmt.Errorf("queryfilter: unknown sort field %q", s.Field)
		}

		if def.SortColumn == "" && def.SortFn == nil {
			return fmt.Errorf("queryfilter: field %q is not sortable", s.Field)
		}
	}

	return nil
}

func opsContain(allowed []Op, op Op) bool {
	return slices.Contains(allowed, op)
}

// ─── GORM scope ──────────────────────────────────────────────────────────────

type dbCond struct {
	sql  string
	args []any
}

func buildGORMScope[T any](node Node, sorts []SortField, schema Schema[T]) func(*gorm.DB) *gorm.DB {
	conds := collectDBConds(node, schema)
	dbSorts := collectDBSorts(sorts, schema)

	if len(conds) == 0 && len(dbSorts) == 0 {
		return nil
	}

	return func(db *gorm.DB) *gorm.DB {
		for _, c := range conds {
			db = db.Where(c.sql, c.args...)
		}

		for _, s := range dbSorts {
			dir := "ASC"
			if s.desc {
				dir = "DESC"
			}

			db = db.Order(fmt.Sprintf("`%s` %s", s.column, dir))
		}

		return db
	}
}

// collectDBConds walks the AST collecting DB-pushable conditions.
// AND: each side collected independently.
// OR: only collected if ALL leaves in the subtree are DB-level.
func collectDBConds[T any](node Node, schema Schema[T]) []dbCond {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ComparisonNode:
		c, ok := leafToDBCond(n, schema)
		if !ok {
			return nil
		}

		return []dbCond{c}

	case *LogicalNode:
		if n.Op == LogicalAND {
			return append(collectDBConds(n.Left, schema), collectDBConds(n.Right, schema)...)
		}

		// OR: only push when entire subtree is DB-level
		c, ok := nodeToDBCond(n, schema)
		if !ok {
			return nil
		}

		return []dbCond{c}
	}

	return nil
}

// nodeToDBCond recursively converts an entire subtree to a single DB condition.
// Returns false if any leaf in the subtree is in-memory.
func nodeToDBCond[T any](node Node, schema Schema[T]) (dbCond, bool) {
	switch n := node.(type) {
	case *ComparisonNode:
		return leafToDBCond(n, schema)

	case *LogicalNode:
		left, lOK := nodeToDBCond(n.Left, schema)
		right, rOK := nodeToDBCond(n.Right, schema)

		if !lOK || !rOK {
			return dbCond{}, false
		}

		op := "AND"
		if n.Op == LogicalOR {
			op = "OR"
		}

		return dbCond{
			sql:  fmt.Sprintf("(%s) %s (%s)", left.sql, op, right.sql),
			args: append(left.args, right.args...),
		}, true
	}

	return dbCond{}, false
}

func leafToDBCond[T any](n *ComparisonNode, schema Schema[T]) (dbCond, bool) {
	def, ok := schema[n.Field]
	if !ok || !def.isDBLevel() {
		return dbCond{}, false
	}

	if def.SQLExpr != nil {
		sql, args, err := def.SQLExpr(n.Op, n.Value)
		if err != nil {
			return dbCond{}, false
		}

		return dbCond{sql: sql, args: args}, true
	}

	sql, args, err := columnCond(def.Column, n.Op, n.Value, def.Type)
	if err != nil {
		return dbCond{}, false
	}

	return dbCond{sql: sql, args: args}, true
}

func columnCond(col, op, val string, typ FieldType) (string, []any, error) {
	q := fmt.Sprintf("`%s`", col)

	switch op {
	case OpEq:
		return fmt.Sprintf("%s = ?", q), []any{coerce(val, typ)}, nil
	case OpNe:
		return fmt.Sprintf("%s != ?", q), []any{coerce(val, typ)}, nil
	case OpGt:
		return fmt.Sprintf("%s > ?", q), []any{coerce(val, typ)}, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", q), []any{coerce(val, typ)}, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", q), []any{coerce(val, typ)}, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", q), []any{coerce(val, typ)}, nil
	case OpPrefix:
		return fmt.Sprintf("%s LIKE ?", q), []any{val + "%"}, nil
	case OpSuffix:
		return fmt.Sprintf("%s LIKE ?", q), []any{"%" + val}, nil
	case OpLike:
		return fmt.Sprintf("%s LIKE ?", q), []any{val}, nil
	default:
		return "", nil, fmt.Errorf("queryfilter: unsupported operator %q for column field", op)
	}
}

func coerce(val string, typ FieldType) any {
	switch typ {
	case FieldTypeInt:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	case FieldTypeBool:
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	case FieldTypeFloat:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}

	return val
}

// ─── Sort helpers ─────────────────────────────────────────────────────────────

type sortDir struct {
	column string
	desc   bool
}

// collectDBSorts returns DB-level sort directives.
// Returns nil when any sort field requires in-memory evaluation (mixed sort),
// deferring all ordering to buildInMemoryFn.
func collectDBSorts[T any](sorts []SortField, schema Schema[T]) []sortDir {
	for _, s := range sorts {
		def, ok := schema[s.Field]
		if ok && def.SortFn != nil {
			return nil // any in-memory sort → all sorts become in-memory
		}
	}

	var dirs []sortDir

	for _, s := range sorts {
		def, ok := schema[s.Field]
		if !ok {
			continue
		}

		col := def.SortColumn
		if col == "" {
			col = def.Column
		}

		if col == "" {
			continue
		}

		dirs = append(dirs, sortDir{column: col, desc: s.Desc})
	}

	return dirs
}

// ─── In-memory filter + sort ──────────────────────────────────────────────────

func buildInMemoryFn[T any](node Node, sorts []SortField, schema Schema[T]) func([]T) []T {
	needFilter := hasMemFilter(node, schema)
	needSort := hasMemSort(sorts, schema)

	if !needFilter && !needSort {
		return nil
	}

	return func(items []T) []T {
		if needFilter {
			filtered := make([]T, 0, len(items))
			for _, item := range items {
				if evalNode(item, node, schema) {
					filtered = append(filtered, item)
				}
			}

			items = filtered
		}

		if needSort {
			sort.SliceStable(items, func(i, j int) bool {
				return compareItems(items[i], items[j], sorts, schema)
			})
		}

		return items
	}
}

func hasMemFilter[T any](node Node, schema Schema[T]) bool {
	if node == nil {
		return false
	}

	switch n := node.(type) {
	case *ComparisonNode:
		def, ok := schema[n.Field]
		return ok && def.FilterFn != nil
	case *LogicalNode:
		return hasMemFilter(n.Left, schema) || hasMemFilter(n.Right, schema)
	}

	return false
}

func hasMemSort[T any](sorts []SortField, schema Schema[T]) bool {
	for _, s := range sorts {
		def, ok := schema[s.Field]
		if ok && def.SortFn != nil {
			return true
		}
	}

	return false
}

// evalNode evaluates the AST against one item.
// DB-level leaves return true (already filtered by the DB query).
// In-memory leaves run their FilterFn.
func evalNode[T any](item T, node Node, schema Schema[T]) bool {
	if node == nil {
		return true
	}

	switch n := node.(type) {
	case *ComparisonNode:
		def, ok := schema[n.Field]
		if !ok {
			return true
		}

		if def.FilterFn == nil {
			return true // DB-level: already handled
		}

		return def.FilterFn(item, n.Op, n.Value)

	case *LogicalNode:
		left := evalNode(item, n.Left, schema)
		right := evalNode(item, n.Right, schema)

		if n.Op == LogicalAND {
			return left && right
		}

		return left || right
	}

	return true
}

// compareItems returns true when a should sort before b.
func compareItems[T any](a, b T, sorts []SortField, schema Schema[T]) bool {
	for _, s := range sorts {
		def, ok := schema[s.Field]
		if !ok || def.SortFn == nil {
			continue
		}

		cmp := def.SortFn(a, b)
		if cmp == 0 {
			continue
		}

		if s.Desc {
			return cmp > 0
		}

		return cmp < 0
	}

	return false
}
