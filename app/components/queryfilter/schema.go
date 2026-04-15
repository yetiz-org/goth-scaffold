package queryfilter

// FieldType represents the data type of a filterable field,
// used for value coercion when building SQL parameters.
type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeInt
	FieldTypeBool
	FieldTypeFloat
)

// FieldDef defines how a query field is filtered and sorted.
//
// Filter path (mutually exclusive — first non-nil wins):
//   - Column   → Level 1: direct DB column, standard SQL operators
//   - SQLExpr  → Level 2: custom SQL fragment (virtual fields, IS NULL checks, mappings)
//   - FilterFn → Level 3: in-memory predicate (fields without a DB column)
//
// Sort path (independent from filter):
//   - SortColumn → ORDER BY a DB column (may differ from Column)
//   - SortFn     → in-memory comparator
//
// Mixed sort note: if any sort field has SortFn, all sorts are evaluated in-memory
// and DB ORDER BY is skipped entirely. Fields that appear in a mixed sort context
// must also provide SortFn to participate in the ordering.
type FieldDef[T any] struct {
	Type       FieldType
	AllowedOps []Op

	// Level 1
	Column string

	// Level 2
	SQLExpr func(op Op, val string) (sql string, args []any, err error)

	// Level 3
	FilterFn func(row T, op Op, val string) bool

	// Sort
	SortColumn string
	SortFn     func(a, b T) int
}

// isDBLevel reports whether this field is processed at DB level for filtering.
func (f FieldDef[T]) isDBLevel() bool {
	return f.FilterFn == nil && (f.Column != "" || f.SQLExpr != nil)
}

// Schema maps RSQL field names to their FieldDef definitions.
type Schema[T any] map[string]FieldDef[T]
