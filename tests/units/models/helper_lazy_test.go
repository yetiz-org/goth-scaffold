package models_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/yetiz-org/goth-scaffold/app/models"
)

// fakeQuerier is a reflection-based stand-in for the GORM-backed querier. Each Find* method
// assigns its pre-seeded slice to dest and records that it was called, so a test can assert
// both the eager-grouping result and that no extra (lazy) query fires afterward.
//
// It implements models.CompositeBatchQuerier (LazyQuerier + BatchQuerier + composite), so it
// can back both the single-FK has-many path (_EagerHasMany) and the scope/composite path.
type fakeQuerier struct {
	rows          any // returned by FindHasMany / FindHasManyIn
	compositeRows any // returned by FindHasManyInComposite

	hasManyCalls        int
	hasManyInCalls      int
	hasManyInComposite  int
	lastCompositeFKCols []string
	lastCompositeExtra  []models.LazyCondition
}

func assignSlice(dest any, rows any) {
	if rows == nil {
		return
	}

	reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(rows))
}

func (f *fakeQuerier) FindById(dest any, id any) error { return nil }

func (f *fakeQuerier) FindHasMany(dest any, fkColumn string, fkValue any) error {
	f.hasManyCalls++
	assignSlice(dest, f.rows)
	return nil
}

func (f *fakeQuerier) FindHasOne(dest any, fkColumn string, fkValue any) error { return nil }

func (f *fakeQuerier) FindByIds(dest any, ids any) error {
	assignSlice(dest, f.rows)
	return nil
}

func (f *fakeQuerier) FindHasManyIn(dest any, fkColumn string, fkValues any) error {
	f.hasManyInCalls++
	assignSlice(dest, f.rows)
	return nil
}

func (f *fakeQuerier) FindHasManyComposite(dest any, conditions map[string]any, extra ...models.LazyCondition) error {
	assignSlice(dest, f.compositeRows)
	return nil
}

func (f *fakeQuerier) FindHasOneComposite(dest any, conditions map[string]any, extra ...models.LazyCondition) error {
	return nil
}

func (f *fakeQuerier) FindHasManyInComposite(dest any, fkColumns []string, compositeKeys [][]any, extra ...models.LazyCondition) error {
	f.hasManyInComposite++
	f.lastCompositeFKCols = fkColumns
	f.lastCompositeExtra = extra
	assignSlice(dest, f.compositeRows)
	return nil
}

// TestEagerTagsPopulatesCache is the core regression guard for the _EagerHasMany rewrite:
// grouping is now keyed by the raw reflect value (SiteSettingId on both the parent reference
// and the child FK). If the two sides did not share the same Go type, grouping would silently
// produce empty caches. The fake returns every seeded tag; correctness means each parent ends
// up with exactly its own tags and no lazy query fires afterward.
func TestEagerTagsPopulatesCache(t *testing.T) {
	tags := []*models.SiteSettingTag{
		{ID: 10, SiteSettingID: 1, Name: "a"},
		{ID: 11, SiteSettingID: 1, Name: "b"},
		{ID: 20, SiteSettingID: 2, Name: "c"},
	}

	fake := &fakeQuerier{rows: tags}
	models.RegisterModelQuerier[models.SiteSettingTag](fake)
	defer models.UnregisterModelQuerier[models.SiteSettingTag]()

	settings := []*models.SiteSetting{{ID: 1}, {ID: 2}, {ID: 3}}
	models.Eager[*models.SiteSetting]("Tags").ApplyEager(settings)

	if fake.hasManyInCalls != 1 {
		t.Fatalf("expected exactly 1 batch IN query, got %d", fake.hasManyInCalls)
	}

	if got := len(settings[0].Tags()); got != 2 {
		t.Errorf("setting 1: expected 2 tags, got %d", got)
	}

	if got := len(settings[1].Tags()); got != 1 {
		t.Errorf("setting 2: expected 1 tag, got %d", got)
	}

	if got := len(settings[2].Tags()); got != 0 {
		t.Errorf("setting 3: expected 0 tags, got %d", got)
	}

	// Tags() after eager must hit the populated cache (incl. the non-nil empty slice for #3),
	// never a fresh query.
	if fake.hasManyCalls != 0 {
		t.Errorf("Tags() after eager should not lazily query, got %d lazy calls", fake.hasManyCalls)
	}

	if fake.hasManyInCalls != 1 {
		t.Errorf("Tags() after eager should not add batch queries, got %d", fake.hasManyInCalls)
	}
}

// TestEagerAllMatchesEager confirms the EagerAll path (allowedFields == nil) still loads the
// has-many association after the _AutoEagerLoad → _AutoEagerLoadFiltered split.
func TestEagerAllMatchesEager(t *testing.T) {
	tags := []*models.SiteSettingTag{
		{ID: 10, SiteSettingID: 1, Name: "a"},
		{ID: 20, SiteSettingID: 2, Name: "c"},
	}

	fake := &fakeQuerier{rows: tags}
	models.RegisterModelQuerier[models.SiteSettingTag](fake)
	defer models.UnregisterModelQuerier[models.SiteSettingTag]()

	settings := []*models.SiteSetting{{ID: 1}, {ID: 2}}
	models.EagerAll[*models.SiteSetting]().ApplyEager(settings)

	if fake.hasManyInCalls != 1 {
		t.Fatalf("EagerAll: expected 1 batch IN query, got %d", fake.hasManyInCalls)
	}

	if got := len(settings[0].Tags()); got != 1 {
		t.Errorf("setting 1: expected 1 tag, got %d", got)
	}

	if got := len(settings[1].Tags()); got != 1 {
		t.Errorf("setting 2: expected 1 tag, got %d", got)
	}
}

// TestEagerSelectiveSkipsUnlistedField confirms Eager(...) only batch-loads the named fields:
// an unknown name leaves the association lazy, so a later Tags() falls back to a single query.
func TestEagerSelectiveSkipsUnlistedField(t *testing.T) {
	tags := []*models.SiteSettingTag{{ID: 10, SiteSettingID: 1, Name: "a"}}

	fake := &fakeQuerier{rows: tags}
	models.RegisterModelQuerier[models.SiteSettingTag](fake)
	defer models.UnregisterModelQuerier[models.SiteSettingTag]()

	settings := []*models.SiteSetting{{ID: 1}}
	models.Eager[*models.SiteSetting]("Bogus").ApplyEager(settings)

	if fake.hasManyInCalls != 0 {
		t.Fatalf("Eager(\"Bogus\") must not batch-load Tags, got %d IN queries", fake.hasManyInCalls)
	}

	// _Tags is still nil → Tags() now triggers a single lazy load.
	_ = settings[0].Tags()
	if fake.hasManyCalls != 1 {
		t.Errorf("expected lazy Tags() to query once, got %d", fake.hasManyCalls)
	}
}

func TestIsZeroDate(t *testing.T) {
	if !models.IsZeroDate(time.Time{}) {
		t.Error("zero time.Time should be reported as a zero date")
	}

	if models.IsZeroDate(time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)) {
		t.Error("a real date should not be reported as a zero date")
	}
}

// ─── scope mechanism (dormant in the scaffold; exercised here via a test-only model) ──────────

type scopeChild struct {
	ID      uint64
	OwnerID uint64
	Channel string
}

func (c *scopeChild) TableName() string { return "scope_children" }

type scopeParent struct {
	ID uint64

	_Children []*scopeChild `gorm:"foreignKey:OwnerID;references:ID" scope:"testChannel"`
}

func (p *scopeParent) SetCacheChildren(v []*scopeChild) { p._Children = v }

func (p *scopeParent) Children() []*scopeChild {
	return models.LazyHasMany[scopeChild](p, &p._Children)
}

// TestScopeRoutesThroughCompositeWithExtraConditions proves the full scope path: a `scope:"..."`
// tag resolves to conditions, which routes the has-many load through the composite querier and
// passes the conditions as the variadic extra arg — even though the FK is single-column.
func TestScopeRoutesThroughCompositeWithExtraConditions(t *testing.T) {
	models.RegisterLazyScope("testChannel", func() []models.LazyCondition {
		return []models.LazyCondition{{Column: "channel", Op: models.LazyOpEq, Value: "A"}}
	})

	children := []*scopeChild{
		{ID: 100, OwnerID: 1, Channel: "A"},
		{ID: 101, OwnerID: 2, Channel: "A"},
	}

	fake := &fakeQuerier{compositeRows: children}
	models.RegisterModelQuerier[scopeChild](fake)
	defer models.UnregisterModelQuerier[scopeChild]()

	parents := []*scopeParent{{ID: 1}, {ID: 2}}
	models.Eager[*scopeParent]("Children").ApplyEager(parents)

	if fake.hasManyInComposite != 1 {
		t.Fatalf("scope tag should route through the composite path once, got %d", fake.hasManyInComposite)
	}

	if fake.hasManyInCalls != 0 {
		t.Errorf("scope tag should NOT use the plain has-many path, got %d", fake.hasManyInCalls)
	}

	if len(fake.lastCompositeExtra) != 1 {
		t.Fatalf("expected 1 scope condition forwarded to querier, got %d", len(fake.lastCompositeExtra))
	}

	cond := fake.lastCompositeExtra[0]
	if cond.Column != "channel" || cond.Op != models.LazyOpEq || cond.Value != "A" {
		t.Errorf("forwarded scope condition mismatch: %+v", cond)
	}

	if got := len(parents[0].Children()); got != 1 {
		t.Errorf("parent 1: expected 1 child, got %d", got)
	}

	if got := len(parents[1].Children()); got != 1 {
		t.Errorf("parent 2: expected 1 child, got %d", got)
	}
}

// TestResolveLazyScopeUnregisteredIsDormant confirms the backward-compat contract: with no
// registration, an unknown scope name resolves to nothing, so association loads keep their
// original (non-scoped) routing.
func TestResolveLazyScopeUnregisteredIsDormant(t *testing.T) {
	tags := []*models.SiteSettingTag{{ID: 10, SiteSettingID: 1, Name: "a"}}

	fake := &fakeQuerier{rows: tags}
	models.RegisterModelQuerier[models.SiteSettingTag](fake)
	defer models.UnregisterModelQuerier[models.SiteSettingTag]()

	// SiteSetting._Tags carries no scope tag → must use the plain has-many path, not composite.
	settings := []*models.SiteSetting{{ID: 1}}
	models.Eager[*models.SiteSetting]("Tags").ApplyEager(settings)

	if fake.hasManyInComposite != 0 {
		t.Errorf("no scope tag should never route through composite, got %d", fake.hasManyInComposite)
	}

	if fake.hasManyInCalls != 1 {
		t.Errorf("expected plain batch IN path, got %d", fake.hasManyInCalls)
	}
}
