package repositories

import (
	"reflect"
	"testing"

	"github.com/yetiz-org/goth-scaffold/app/models"
)

// TestRenderLazyCond verifies the SQL fragment produced for each LazyCondition operator.
// This is the backported scope-rendering logic; it is dormant until a downstream registers a
// scope, so it has no live caller in the scaffold and is covered only here. The column is
// quoted via the active dialect, which defaults to MySQL (backticks) when no DB is initialised.
func TestRenderLazyCond(t *testing.T) {
	const col = "`channel`" // _QuoteSQLIdentifier("channel") under the default MySQL dialect

	cases := []struct {
		op         string
		wantClause string
		wantArgs   []any
	}{
		{models.LazyOpEq, col + " = ?", []any{"v"}},
		{models.LazyOpNe, col + " != ?", []any{"v"}},
		{models.LazyOpGt, col + " > ?", []any{"v"}},
		{models.LazyOpGte, col + " >= ?", []any{"v"}},
		{models.LazyOpLt, col + " < ?", []any{"v"}},
		{models.LazyOpLte, col + " <= ?", []any{"v"}},
		{models.LazyOpLike, col + " LIKE ?", []any{"v"}},
		{models.LazyOpPrefix, col + " LIKE ?", []any{"v%"}},
		{models.LazyOpSuffix, col + " LIKE ?", []any{"%v"}},
		{models.LazyOpEmpty, "COALESCE(" + col + ", '') = ''", nil},
		{models.LazyOpNotEmpty, "COALESCE(" + col + ", '') != ''", nil},
	}

	for _, tc := range cases {
		t.Run(tc.op, func(t *testing.T) {
			clause, args, ok := _RenderLazyCond(models.LazyCondition{Column: "channel", Op: tc.op, Value: "v"})
			if !ok {
				t.Fatalf("op %q: expected ok=true", tc.op)
			}

			if clause != tc.wantClause {
				t.Errorf("op %q: clause = %q, want %q", tc.op, clause, tc.wantClause)
			}

			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Errorf("op %q: args = %#v, want %#v", tc.op, args, tc.wantArgs)
			}
		})
	}
}

func TestRenderLazyCondUnsupportedOp(t *testing.T) {
	clause, args, ok := _RenderLazyCond(models.LazyCondition{Column: "channel", Op: "bogus", Value: "v"})
	if ok {
		t.Fatalf("unsupported op should return ok=false, got clause=%q", clause)
	}

	if clause != "" || args != nil {
		t.Errorf("unsupported op should return empty clause/args, got %q / %#v", clause, args)
	}
}
