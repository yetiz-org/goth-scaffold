package queryfilter

import (
	"strings"
	"testing"
)

// ─── ParseFilter tests ────────────────────────────────────────────────────────

func TestParseFilterEmptyReturnsNil(t *testing.T) {
	node, err := ParseFilter("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if node != nil {
		t.Fatalf("expected nil node for empty input, got %T", node)
	}
}

func TestParseFilterSimpleEq(t *testing.T) {
	node, err := ParseFilter("name==foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := node.(*ComparisonNode)
	if !ok {
		t.Fatalf("expected ComparisonNode, got %T", node)
	}
	if c.Field != "name" || c.Op != OpEq || c.Value != "foo" {
		t.Errorf("got field=%q op=%q val=%q", c.Field, c.Op, c.Value)
	}
}

func TestParseFilterNamedOpEq(t *testing.T) {
	node, err := ParseFilter("age=eq=25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := node.(*ComparisonNode)
	if !ok {
		t.Fatalf("expected ComparisonNode, got %T", node)
	}
	if c.Field != "age" || c.Op != OpEq || c.Value != "25" {
		t.Errorf("got field=%q op=%q val=%q", c.Field, c.Op, c.Value)
	}
}

func TestParseFilterAndExpression(t *testing.T) {
	node, err := ParseFilter("name==foo;age=gt=10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l, ok := node.(*LogicalNode)
	if !ok {
		t.Fatalf("expected LogicalNode, got %T", node)
	}
	if l.Op != LogicalAND {
		t.Errorf("expected AND, got %q", l.Op)
	}
}

func TestParseFilterOrExpression(t *testing.T) {
	node, err := ParseFilter("status==active,status==pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l, ok := node.(*LogicalNode)
	if !ok {
		t.Fatalf("expected LogicalNode, got %T", node)
	}
	if l.Op != LogicalOR {
		t.Errorf("expected OR, got %q", l.Op)
	}
}

func TestParseFilterParentheses(t *testing.T) {
	node, err := ParseFilter("(name==foo,name==bar);active==true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	l, ok := node.(*LogicalNode)
	if !ok {
		t.Fatalf("expected LogicalNode, got %T", node)
	}
	if l.Op != LogicalAND {
		t.Errorf("expected AND at top level, got %q", l.Op)
	}
}

func TestParseFilterQuotedString(t *testing.T) {
	node, err := ParseFilter(`name=="hello world"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := node.(*ComparisonNode)
	if !ok {
		t.Fatalf("expected ComparisonNode, got %T", node)
	}
	if c.Value != "hello world" {
		t.Errorf("expected value %q, got %q", "hello world", c.Value)
	}
}

func TestParseFilterAllComparisonOps(t *testing.T) {
	cases := []struct {
		expr string
		op   Op
	}{
		{"f==v", OpEq},
		{"f!=v", OpNe},
		{"f>v", OpGt},
		{"f<v", OpLt},
		{"f>=v", OpGte},
		{"f<=v", OpLte},
		{"f=prefix=v", OpPrefix},
		{"f=suffix=v", OpSuffix},
		{"f=like=v", OpLike},
	}
	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			node, err := ParseFilter(tc.expr)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.expr, err)
			}
			c, ok := node.(*ComparisonNode)
			if !ok {
				t.Fatalf("expected ComparisonNode, got %T", node)
			}
			if c.Op != tc.op {
				t.Errorf("expected op %q, got %q", tc.op, c.Op)
			}
		})
	}
}

func TestParseFilterInvalidReturnsError(t *testing.T) {
	cases := []string{
		"@invalid",
		"name=unknownop=val",
		"(name==foo",
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, err := ParseFilter(expr)
			if err == nil {
				t.Errorf("expected error for %q, got nil", expr)
			}
		})
	}
}

// ─── ParseSort tests ──────────────────────────────────────────────────────────

func TestParseSortEmptyReturnsNil(t *testing.T) {
	result := ParseSort("")
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}
}

func TestParseSortAscending(t *testing.T) {
	result := ParseSort("name")
	if len(result) != 1 {
		t.Fatalf("expected 1 field, got %d", len(result))
	}
	if result[0].Field != "name" || result[0].Desc {
		t.Errorf("expected asc name, got %+v", result[0])
	}
}

func TestParseSortDescending(t *testing.T) {
	result := ParseSort("-name")
	if len(result) != 1 {
		t.Fatalf("expected 1 field, got %d", len(result))
	}
	if result[0].Field != "name" || !result[0].Desc {
		t.Errorf("expected desc name, got %+v", result[0])
	}
}

func TestParseSortPlusPrefix(t *testing.T) {
	result := ParseSort("+name")
	if len(result) != 1 || result[0].Desc {
		t.Errorf("expected asc name, got %+v", result)
	}
}

func TestParseSortMultipleFields(t *testing.T) {
	result := ParseSort("+year,-title,status")
	if len(result) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(result), result)
	}
	if result[0].Field != "year" || result[0].Desc {
		t.Errorf("field[0] = %+v", result[0])
	}
	if result[1].Field != "title" || !result[1].Desc {
		t.Errorf("field[1] = %+v", result[1])
	}
	if result[2].Field != "status" || result[2].Desc {
		t.Errorf("field[2] = %+v", result[2])
	}
}

// ─── ValidateNode tests ───────────────────────────────────────────────────────

type testRow struct {
	Name  string
	Age   int
	Score float64
}

func testSchema() Schema[testRow] {
	return Schema[testRow]{
		"name": {
			Type:       FieldTypeString,
			Column:     "name",
			SortColumn: "name",
		},
		"age": {
			Type:       FieldTypeInt,
			Column:     "age",
			AllowedOps: []Op{OpEq, OpGt, OpLt},
		},
		"score": {
			Type:   FieldTypeFloat,
			Column: "score",
		},
		"active": {
			Type:   FieldTypeBool,
			Column: "active",
		},
		"memo": {
			// in-memory only
			FilterFn: func(row testRow, op Op, val string) bool {
				return strings.Contains(row.Name, val)
			},
		},
	}
}

func TestValidateNodeNilReturnsNil(t *testing.T) {
	if err := ValidateNode[testRow](nil, testSchema()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateNodeUnknownFieldReturnsError(t *testing.T) {
	node := &ComparisonNode{Field: "unknown", Op: OpEq, Value: "x"}
	err := ValidateNode(node, testSchema())
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNodeDisallowedOpReturnsError(t *testing.T) {
	// age only allows eq, gt, lt
	node := &ComparisonNode{Field: "age", Op: OpPrefix, Value: "10"}
	err := ValidateNode(node, testSchema())
	if err == nil {
		t.Fatal("expected error for disallowed op")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNodeIntFieldWithNonIntValueReturnsError(t *testing.T) {
	node := &ComparisonNode{Field: "age", Op: OpEq, Value: "notanint"}
	err := ValidateNode(node, testSchema())
	if err == nil {
		t.Fatal("expected error for non-int value on int field")
	}
}

func TestValidateNodeValidIntPasses(t *testing.T) {
	node := &ComparisonNode{Field: "age", Op: OpEq, Value: "42"}
	if err := ValidateNode(node, testSchema()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── ValidateSorts tests ──────────────────────────────────────────────────────

func TestValidateSortsUnknownFieldReturnsError(t *testing.T) {
	sorts := []SortField{{Field: "unknown"}}
	err := ValidateSorts(sorts, testSchema())
	if err == nil {
		t.Fatal("expected error for unknown sort field")
	}
}

func TestValidateSortsUnsortableFieldReturnsError(t *testing.T) {
	// memo has no SortColumn and no SortFn
	sorts := []SortField{{Field: "memo"}}
	err := ValidateSorts(sorts, testSchema())
	if err == nil {
		t.Fatal("expected error for non-sortable field")
	}
}

func TestValidateSortsValidFieldPasses(t *testing.T) {
	sorts := []SortField{{Field: "name"}}
	if err := ValidateSorts(sorts, testSchema()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── In-memory filter tests (via ToOption._FilterFn) ─────────────────────────

func applyInMemoryFilter(t *testing.T, filterExpr string, items []testRow) []testRow {
	t.Helper()
	node, err := ParseFilter(filterExpr)
	if err != nil {
		t.Fatalf("ParseFilter(%q) error: %v", filterExpr, err)
	}
	schema := testSchema()
	if err := ValidateNode(node, schema); err != nil {
		t.Fatalf("ValidateNode error: %v", err)
	}
	filterFn := buildInMemoryFn(node, nil, schema)
	if filterFn == nil {
		return items
	}
	return filterFn(items)
}

func TestInMemoryFilterMemoContains(t *testing.T) {
	items := []testRow{
		{Name: "alice"},
		{Name: "bob"},
		{Name: "alice_extra"},
	}
	result := applyInMemoryFilter(t, "memo==alice", items)
	if len(result) != 2 {
		t.Errorf("expected 2 items containing 'alice', got %d: %v", len(result), result)
	}
}

func TestInMemoryFilterMemoAndInMemory(t *testing.T) {
	items := []testRow{
		{Name: "alice", Age: 30},
		{Name: "bob", Age: 25},
		{Name: "alice_young", Age: 20},
	}
	// memo (in-memory) AND name (DB-level, passes through)
	result := applyInMemoryFilter(t, "memo==alice", items)
	for _, r := range result {
		if !strings.Contains(r.Name, "alice") {
			t.Errorf("unexpected item in result: %v", r)
		}
	}
}

// ─── In-memory sort tests ─────────────────────────────────────────────────────

func TestInMemorySortByStringFn(t *testing.T) {
	schema := Schema[testRow]{
		"name": {
			SortFn: func(a, b testRow) int {
				if a.Name < b.Name {
					return -1
				}
				if a.Name > b.Name {
					return 1
				}
				return 0
			},
		},
	}

	items := []testRow{{Name: "charlie"}, {Name: "alice"}, {Name: "bob"}}
	sorts := []SortField{{Field: "name", Desc: false}}

	filterFn := buildInMemoryFn[testRow](nil, sorts, schema)
	if filterFn == nil {
		t.Fatal("expected non-nil filterFn")
	}
	result := filterFn(items)
	if result[0].Name != "alice" || result[1].Name != "bob" || result[2].Name != "charlie" {
		t.Errorf("unexpected sort order: %v", result)
	}
}

func TestInMemorySortDescending(t *testing.T) {
	schema := Schema[testRow]{
		"age": {
			SortFn: func(a, b testRow) int {
				return a.Age - b.Age
			},
		},
	}

	items := []testRow{{Age: 10}, {Age: 30}, {Age: 20}}
	sorts := []SortField{{Field: "age", Desc: true}}

	filterFn := buildInMemoryFn[testRow](nil, sorts, schema)
	result := filterFn(items)
	if result[0].Age != 30 || result[1].Age != 20 || result[2].Age != 10 {
		t.Errorf("unexpected desc sort: %v", result)
	}
}

// ─── opsContain tests ─────────────────────────────────────────────────────────

func TestOpsContainReturnsTrueWhenPresent(t *testing.T) {
	if !opsContain([]Op{OpEq, OpGt}, OpGt) {
		t.Error("expected opsContain to return true")
	}
}

func TestOpsContainReturnsFalseWhenAbsent(t *testing.T) {
	if opsContain([]Op{OpEq, OpGt}, OpLt) {
		t.Error("expected opsContain to return false")
	}
}

// ─── fieldDef.isDBLevel tests ─────────────────────────────────────────────────

func TestFieldDefIsDBLevelWithColumn(t *testing.T) {
	def := FieldDef[testRow]{Column: "my_col"}
	if !def.isDBLevel() {
		t.Error("expected isDBLevel=true for Column-only field")
	}
}

func TestFieldDefIsDBLevelWithFilterFn(t *testing.T) {
	def := FieldDef[testRow]{
		Column:   "my_col",
		FilterFn: func(row testRow, op Op, val string) bool { return true },
	}
	if def.isDBLevel() {
		t.Error("expected isDBLevel=false when FilterFn is set")
	}
}

func TestFieldDefIsDBLevelWithNoColumn(t *testing.T) {
	def := FieldDef[testRow]{}
	if def.isDBLevel() {
		t.Error("expected isDBLevel=false when Column and SQLExpr are empty")
	}
}
