package models

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	kklogger "github.com/yetiz-org/goth-kklogger"
)

// LazyQuerier is the interface that the repository layer must implement
// to execute database queries on behalf of lazy-loading helpers in this package.
// Repositories inject their implementation via RegisterLazyQuerier / RegisterModelQuerier.
type LazyQuerier interface {
	// FindById loads a single record whose primary key equals id into dest.
	FindById(dest any, id any) error
	// FindHasMany loads all records where fkColumn = fkValue into dest (must be a pointer to a slice).
	FindHasMany(dest any, fkColumn string, fkValue any) error
	// FindHasOne loads the first record where fkColumn = fkValue into dest.
	FindHasOne(dest any, fkColumn string, fkValue any) error
}

// _LazyMeta holds parsed association tag metadata, cached per model+field.
type _LazyMeta struct {
	FKFieldName  string
	RefFieldName string
	FKIsPtr      bool

	FKFieldNames  []string
	RefFieldNames []string

	ScopeName string // `scope:"<name>"` — resolved at load time via _ResolveLazyScope
}

func (m *_LazyMeta) _IsComposite() bool { return len(m.FKFieldNames) > 1 }

var _LazyMetaCache sync.Map

// _LazyDefaultQuerier is the fallback querier used when no type-specific override exists.
var _LazyDefaultQuerier LazyQuerier

// _LazyQuerierRegistry maps reflect.Type → LazyQuerier for per-model DB routing.
var _LazyQuerierRegistry sync.Map

// _GetQuerierForType returns the querier for the given model type.
func _GetQuerierForType(t reflect.Type) LazyQuerier {
	if q, ok := _LazyQuerierRegistry.Load(t); ok {
		return q.(LazyQuerier)
	}

	return _LazyDefaultQuerier
}

// GetDefaultQuerier returns the currently registered default LazyQuerier.
func GetDefaultQuerier() LazyQuerier {
	return _LazyDefaultQuerier
}

// RegisterLazyQuerier registers the default querier for lazy loading.
// Call once at application startup (typically in a repositories init()).
func RegisterLazyQuerier(q LazyQuerier) {
	_LazyDefaultQuerier = q
}

// RegisterModelQuerier registers a type-specific querier for T.
func RegisterModelQuerier[T any](q LazyQuerier) {
	t := reflect.TypeFor[T]()
	_LazyQuerierRegistry.Store(t, q)
}

// UnregisterModelQuerier removes the type-specific querier override for T (mainly for tests).
func UnregisterModelQuerier[T any]() {
	t := reflect.TypeFor[T]()
	_LazyQuerierRegistry.Delete(t)
}

func _GetLazyMeta(modelType reflect.Type, fieldName string) *_LazyMeta {
	cacheKey := modelType.Name() + "." + fieldName
	if v, ok := _LazyMetaCache.Load(cacheKey); ok {
		return v.(*_LazyMeta)
	}

	field, ok := modelType.FieldByName(fieldName)
	if !ok {
		panic("models: lazy field not found: " + modelType.Name() + "." + fieldName)
	}

	gormTag := field.Tag.Get("gorm")

	fkRaw := _ParseGormTagValue(gormTag, "foreignKey")
	if fkRaw == "" {
		panic("models: lazy field " + modelType.Name() + "." + fieldName + " missing foreignKey in gorm tag")
	}

	fkNames := _SplitCompositeTag(fkRaw)

	refRaw := _ParseGormTagValue(gormTag, "references")
	var refNames []string
	if refRaw == "" {
		refNames = []string{"ID"}
	} else {
		refNames = _SplitCompositeTag(refRaw)
	}

	meta := &_LazyMeta{
		FKFieldName:   fkNames[0],
		RefFieldName:  refNames[0],
		FKFieldNames:  fkNames,
		RefFieldNames: refNames,
		ScopeName:     field.Tag.Get("scope"),
	}

	// FKIsPtr only meaningful for single FK (belongs-to).
	if !meta._IsComposite() {
		if fkField, ok := modelType.FieldByName(meta.FKFieldName); ok {
			meta.FKIsPtr = fkField.Type.Kind() == reflect.Pointer
		}
	}

	_LazyMetaCache.Store(cacheKey, meta)

	return meta
}

func _ParseGormTagValue(tag, key string) string {
	for part := range strings.SplitSeq(tag, ";") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == key {
			return strings.TrimSpace(kv[1])
		}
	}

	return ""
}

// _SplitCompositeTag splits a comma-separated GORM tag value into trimmed field names.
func _SplitCompositeTag(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

// ResolveBelongsTo reads the gorm foreignKey tag from the named field,
// extracts the FK value via reflect, and calls Resolve() on it.
func ResolveBelongsTo[T any](model any, fieldName string) *T {
	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Pointer {
		modelVal = modelVal.Elem()
	}

	meta := _GetLazyMeta(modelVal.Type(), fieldName)

	fkField := modelVal.FieldByName(meta.FKFieldName)

	if meta.FKIsPtr {
		if fkField.IsNil() {
			return nil
		}

		fkField = fkField.Elem()
	}

	resolveMethod := fkField.MethodByName("Resolve")
	if !resolveMethod.IsValid() {
		return nil
	}

	results := resolveMethod.Call(nil)
	if results[0].IsNil() {
		return nil
	}

	return results[0].Interface().(*T)
}

// ResolveById loads a single model by primary key via the registered LazyQuerier.
func ResolveById[T any](id any) *T {
	if id == nil {
		return nil
	}

	if reflect.ValueOf(id).IsZero() {
		return nil
	}

	childType := reflect.TypeFor[T]()
	q := _GetQuerierForType(childType)
	if q == nil {
		return nil
	}

	result := new(T)
	if err := q.FindById(result, id); err != nil {
		return nil
	}

	return result
}

// _ChildColumn returns the database column name for fieldName on childType.
func _ChildColumn(childType reflect.Type, fieldName string) string {
	if field, ok := childType.FieldByName(fieldName); ok {
		if col := _ParseGormTagValue(field.Tag.Get("gorm"), "column"); col != "" {
			return col
		}
	}

	return _ToSnakeCase(fieldName)
}

// LazyHasMany returns the cached has-many slice for field, loading it on first access.
func LazyHasMany[T any](model any, field *[]*T, opts ...DatabaseQueryOption[*T]) []*T {
	if *field == nil {
		modelVal := reflect.ValueOf(model)
		if modelVal.Kind() == reflect.Pointer {
			modelVal = modelVal.Elem()
		}

		fieldName := _FindFieldByAddr(modelVal, reflect.ValueOf(field).Pointer())
		if fieldName == "" {
			panic("models: LazyHasMany: field address not found in " + modelVal.Type().Name())
		}

		meta := _GetLazyMeta(modelVal.Type(), fieldName)
		childType := reflect.TypeFor[T]()
		scope := _ResolveLazyScope(meta.ScopeName)

		conditions, ok := _BuildFKConditions(modelVal, meta, childType)
		if !ok {
			*field = []*T{}
		} else {
			var results []*T
			if meta._IsComposite() || len(scope) > 0 {
				cq := _GetCompositeQuerierForType(childType)
				if cq == nil {
					return []*T{}
				}

				if err := cq.FindHasManyComposite(&results, conditions, scope...); err != nil {
					kklogger.ErrorJ("models:LazyHasMany#composite_query", err.Error())
					return []*T{}
				}
			} else {
				q := _GetQuerierForType(childType)
				if q == nil {
					return []*T{}
				}

				fkColumn := _ChildColumn(childType, meta.FKFieldName)
				if err := q.FindHasMany(&results, fkColumn, conditions[fkColumn]); err != nil {
					kklogger.ErrorJ("models:LazyHasMany#query", err.Error())
					return []*T{}
				}
			}

			*field = results
		}
	}

	results := *field
	for _, opt := range opts {
		opt.ApplyEager(results)
	}

	for _, opt := range opts {
		results = opt.ApplyFilter(results)
	}

	return results
}

// LazyHasOne returns the cached has-one record for field, loading it on first access.
func LazyHasOne[T any](model any, field **T, opts ...DatabaseQueryOption[*T]) *T {
	if *field == nil {
		modelVal := reflect.ValueOf(model)
		if modelVal.Kind() == reflect.Pointer {
			modelVal = modelVal.Elem()
		}

		fieldName := _FindFieldByAddr(modelVal, reflect.ValueOf(field).Pointer())
		if fieldName == "" {
			panic("models: LazyHasOne: field address not found in " + modelVal.Type().Name())
		}

		meta := _GetLazyMeta(modelVal.Type(), fieldName)
		childType := reflect.TypeFor[T]()
		scope := _ResolveLazyScope(meta.ScopeName)

		conditions, ok := _BuildFKConditions(modelVal, meta, childType)
		if !ok {
			return nil
		}

		var result *T
		if meta._IsComposite() || len(scope) > 0 {
			cq := _GetCompositeQuerierForType(childType)
			if cq == nil {
				return nil
			}

			result = new(T)
			if err := cq.FindHasOneComposite(result, conditions, scope...); err != nil {
				kklogger.ErrorJ("models:LazyHasOne#composite_query", err.Error())
				return nil
			}
		} else {
			q := _GetQuerierForType(childType)
			if q == nil {
				return nil
			}

			fkColumn := _ChildColumn(childType, meta.FKFieldName)
			result = new(T)
			if err := q.FindHasOne(result, fkColumn, conditions[fkColumn]); err != nil {
				return nil
			}
		}

		*field = result
	}

	if *field != nil {
		for _, opt := range opts {
			opt.ApplyEager([]*T{*field})
		}
	}

	return *field
}

// LazyBelongsTo returns the cached belongs-to record for field, loading it on first access.
func LazyBelongsTo[T any](model any, field **T, opts ...DatabaseQueryOption[*T]) *T {
	if *field == nil {
		modelVal := reflect.ValueOf(model)
		if modelVal.Kind() == reflect.Pointer {
			modelVal = modelVal.Elem()
		}

		fieldName := _FindFieldByAddr(modelVal, reflect.ValueOf(field).Pointer())
		if fieldName == "" {
			panic("models: LazyBelongsTo: field address not found in " + modelVal.Type().Name())
		}

		result := ResolveBelongsTo[T](model, fieldName)
		if result != nil {
			*field = result
		}
	}

	if *field != nil {
		for _, opt := range opts {
			opt.ApplyEager([]*T{*field})
		}
	}

	return *field
}

// _BuildFKConditions resolves the parent's reference values into a child-column→value
// map. Returns (nil, false) when any required reference is a nil pointer so the caller
// can short-circuit to an empty result.
func _BuildFKConditions(modelVal reflect.Value, meta *_LazyMeta, childType reflect.Type) (map[string]any, bool) {
	conditions := make(map[string]any, len(meta.FKFieldNames))
	for i, fkName := range meta.FKFieldNames {
		refField := modelVal.FieldByName(meta.RefFieldNames[i])
		if refField.Kind() == reflect.Pointer {
			if refField.IsNil() {
				return nil, false
			}

			refField = refField.Elem()
		}

		conditions[_ChildColumn(childType, fkName)] = refField.Interface()
	}

	return conditions, true
}

// _FindFieldByAddr finds the field name in modelVal whose address matches targetAddr.
func _FindFieldByAddr(modelVal reflect.Value, targetAddr uintptr) string {
	for i := 0; i < modelVal.NumField(); i++ {
		f := modelVal.Field(i)
		if !f.CanAddr() {
			continue
		}

		if f.Addr().Pointer() == targetAddr {
			return modelVal.Type().Field(i).Name
		}
	}

	return ""
}

// ResolveHasMany reads the gorm foreignKey tag from the named field,
// extracts the parent PK value, and queries the child table via the registered LazyQuerier.
func ResolveHasMany[T any](model any, fieldName string) []*T {
	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Pointer {
		modelVal = modelVal.Elem()
	}

	meta := _GetLazyMeta(modelVal.Type(), fieldName)
	childType := reflect.TypeFor[T]()

	var results []*T

	if meta._IsComposite() {
		cq := _GetCompositeQuerierForType(childType)
		if cq == nil {
			return []*T{}
		}

		conditions := make(map[string]any, len(meta.FKFieldNames))
		for i, fkName := range meta.FKFieldNames {
			col := _ChildColumn(childType, fkName)
			refField := modelVal.FieldByName(meta.RefFieldNames[i])
			if refField.Kind() == reflect.Pointer {
				if refField.IsNil() {
					return []*T{}
				}
				refField = refField.Elem()
			}
			conditions[col] = refField.Interface()
		}

		if err := cq.FindHasManyComposite(&results, conditions); err != nil {
			kklogger.ErrorJ("models:ResolveHasMany#composite_query", err.Error())
			return []*T{}
		}
	} else {
		refField := modelVal.FieldByName(meta.RefFieldName)
		pkValue := refField.Interface()

		fkColumn := _ChildColumn(childType, meta.FKFieldName)

		q := _GetQuerierForType(childType)
		if q == nil {
			return []*T{}
		}

		if err := q.FindHasMany(&results, fkColumn, pkValue); err != nil {
			kklogger.ErrorJ("models:ResolveHasMany#query", err.Error())
			return []*T{}
		}
	}

	return results
}

// _ToSnakeCase converts a CamelCase identifier to snake_case.
// Consecutive uppercase letters (e.g. "ID", "URL") are treated as a single word:
//
//	FieldID  → field_id
//	ABCTest  → abc_test
func _ToSnakeCase(s string) string {
	n := len(s)
	result := make([]byte, 0, n+4)

	for i := range n {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				prevIsLower := s[i-1] >= 'a' && s[i-1] <= 'z'
				nextIsLower := i+1 < n && s[i+1] >= 'a' && s[i+1] <= 'z'
				if prevIsLower || (nextIsLower && s[i-1] >= 'A' && s[i-1] <= 'Z') {
					result = append(result, '_')
				}
			}

			result = append(result, c+32)
		} else {
			result = append(result, c)
		}
	}

	return string(result)
}

// ─── Batch Eager Loading ──────────────────────────────────────────────────────

// BatchQuerier extends LazyQuerier with batch IN-query capabilities.
type BatchQuerier interface {
	LazyQuerier
	FindByIds(dest any, ids any) error
	FindHasManyIn(dest any, fkColumn string, fkValues any) error
}

func _GetBatchQuerierForType(t reflect.Type) BatchQuerier {
	if bq, ok := _GetQuerierForType(t).(BatchQuerier); ok {
		return bq
	}

	return nil
}

// CompositeQuerier extends LazyQuerier with composite-FK support. conditions is an
// AND-equality FK match; extra carries scope conditions (`scope:"<name>"`) applied as
// additional WHERE terms (any operator, AND-joined).
type CompositeQuerier interface {
	LazyQuerier
	FindHasManyComposite(dest any, conditions map[string]any, extra ...LazyCondition) error
	FindHasOneComposite(dest any, conditions map[string]any, extra ...LazyCondition) error
}

// CompositeBatchQuerier extends BatchQuerier with `(fkColumns) IN compositeKeys` support.
// extra carries scope conditions (any operator, AND-joined). Single-FK fields with a scope
// route through this path with a one-column tuple.
type CompositeBatchQuerier interface {
	BatchQuerier
	FindHasManyInComposite(dest any, fkColumns []string, compositeKeys [][]any, extra ...LazyCondition) error
}

func _GetCompositeQuerierForType(t reflect.Type) CompositeQuerier {
	if cq, ok := _GetQuerierForType(t).(CompositeQuerier); ok {
		return cq
	}

	return nil
}

func _GetCompositeBatchQuerierForType(t reflect.Type) CompositeBatchQuerier {
	if cbq, ok := _GetQuerierForType(t).(CompositeBatchQuerier); ok {
		return cbq
	}

	return nil
}

// _AutoEagerLoad is the reflect-based entry point called by EagerAll[T]().
// T must be a pointer model type (e.g. *SiteSetting).
// Scans all unexported struct fields with gorm foreignKey tags and issues
// batch IN queries per association instead of per-record lazy loads.
func _AutoEagerLoad[T any](items []T) {
	_AutoEagerLoadFiltered(items, nil)
}

// _AutoEagerLoadFiltered is the selective variant of _AutoEagerLoad. When allowedFields
// is non-nil, only lazy fields whose normalized name (struct field name with the leading
// "_" stripped) is present in allowedFields are batch-loaded. Passing nil preserves the
// original "load every lazy field" behavior of EagerAll.
//
// Callers should not invoke this directly; use models.Eager[T](names...) or
// models.EagerAll[T]() to build the corresponding DatabaseQueryOption.
func _AutoEagerLoadFiltered[T any](items []T, allowedFields map[string]struct{}) {
	if len(items) == 0 {
		return
	}

	firstPtr := reflect.ValueOf(items[0])
	if firstPtr.Kind() != reflect.Pointer {
		return
	}

	modelType := firstPtr.Type().Elem()

	rvItems := make([]reflect.Value, len(items))
	for i, item := range items {
		rvItems[i] = reflect.ValueOf(item)
	}

	for field := range modelType.Fields() {
		if field.IsExported() {
			continue
		}

		fkFieldName := _ParseGormTagValue(field.Tag.Get("gorm"), "foreignKey")
		if fkFieldName == "" {
			continue
		}

		if allowedFields != nil {
			normalized := strings.TrimPrefix(field.Name, "_")
			if _, ok := allowedFields[normalized]; !ok {
				continue
			}
		}

		refFieldName := _ParseGormTagValue(field.Tag.Get("gorm"), "references")

		switch field.Type.Kind() {
		case reflect.Pointer:
			fkNames := _SplitCompositeTag(fkFieldName)
			isBelongsTo := true
			for _, fk := range fkNames {
				if _, ok := modelType.FieldByName(fk); !ok {
					isBelongsTo = false
					break
				}
			}
			if isBelongsTo {
				_EagerBelongsTo(rvItems, field, fkFieldName, field.Type.Elem())
			}
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Pointer {
				scope := _ResolveLazyScope(field.Tag.Get("scope"))
				if strings.Contains(fkFieldName, ",") || len(scope) > 0 {
					_EagerHasManyComposite(rvItems, field, fkFieldName, refFieldName, field.Type.Elem().Elem(), scope)
				} else {
					_EagerHasMany(rvItems, field, fkFieldName, refFieldName, field.Type.Elem().Elem())
				}
			}
		}
	}
}

// _EagerBelongsTo batch-loads a belongs-to association.
func _EagerBelongsTo(items []reflect.Value, field reflect.StructField, fkFieldName string, childType reflect.Type) {
	bq := _GetBatchQuerierForType(childType)
	if bq == nil {
		return
	}

	modelType := items[0].Type().Elem()
	fkStructField, ok := modelType.FieldByName(fkFieldName)
	if !ok {
		return
	}

	fkIsPtr := fkStructField.Type.Kind() == reflect.Pointer

	fkToItems := make(map[uint64][]reflect.Value)
	for _, rv := range items {
		mv := rv.Elem()
		if !mv.FieldByName(field.Name).IsNil() {
			continue
		}

		fkVal := mv.FieldByName(fkFieldName)
		if fkIsPtr {
			if fkVal.IsNil() {
				continue
			}

			fkVal = fkVal.Elem()
		}

		id := _ExtractUint64(fkVal)
		if id == 0 {
			continue
		}

		fkToItems[id] = append(fkToItems[id], rv)
	}

	if len(fkToItems) == 0 {
		return
	}

	ids := _MapKeys(fkToItems)
	resultSlice := reflect.New(reflect.SliceOf(reflect.PointerTo(childType))).Elem()
	if err := bq.FindByIds(resultSlice.Addr().Interface(), ids); err != nil {
		kklogger.ErrorJ("models:_EagerBelongsTo#query", err.Error())
		return
	}

	childByID := make(map[uint64]reflect.Value, resultSlice.Len())
	for i := 0; i < resultSlice.Len(); i++ {
		c := resultSlice.Index(i)
		childByID[_ExtractUint64(c.Elem().FieldByName("ID"))] = c
	}

	setCacheName := "SetCache" + strings.TrimPrefix(field.Name, "_")
	nilVal := reflect.Zero(field.Type)
	for fkID, rvs := range fkToItems {
		val := nilVal
		if c, ok := childByID[fkID]; ok {
			val = c
		}

		for _, rv := range rvs {
			if m := rv.MethodByName(setCacheName); m.IsValid() {
				m.Call([]reflect.Value{val})
			}
		}
	}
}

// _EagerHasMany batch-loads a has-many association keyed by the parent's reference field.
// refFieldName falls back to "ID" when empty (matching gorm's default references=primary key),
// so callers can support non-PK references (e.g. a string column matched against a string FK).
// Items whose cache slice is already non-nil are skipped.
// Items with no children receive an empty (non-nil) slice to prevent future lazy loads.
//
// Grouping is keyed by the raw reflect value (reference value on the parent, FK value on the
// child). Both sides must share the same Go type for the keys to match — which holds by the
// app/models convention that an FK field uses the referenced model's <Model>Id type.
func _EagerHasMany(items []reflect.Value, field reflect.StructField, fkFieldName, refFieldName string, childType reflect.Type) {
	bq := _GetBatchQuerierForType(childType)
	if bq == nil {
		return
	}

	if refFieldName == "" {
		refFieldName = "ID"
	}

	fkColumn := _ChildColumn(childType, fkFieldName)
	sliceType := reflect.SliceOf(reflect.PointerTo(childType))

	keyToItems := make(map[any][]reflect.Value)
	parentKeys := make([]any, 0)
	for _, rv := range items {
		mv := rv.Elem()
		if !mv.FieldByName(field.Name).IsNil() {
			continue
		}

		refField := mv.FieldByName(refFieldName)
		if !refField.IsValid() {
			continue
		}

		if refField.Kind() == reflect.Pointer {
			if refField.IsNil() {
				continue
			}

			refField = refField.Elem()
		}

		if refField.IsZero() {
			continue
		}

		key := refField.Interface()
		if _, exists := keyToItems[key]; !exists {
			parentKeys = append(parentKeys, key)
		}

		keyToItems[key] = append(keyToItems[key], rv)
	}

	if len(keyToItems) == 0 {
		return
	}

	resultSlice := reflect.New(sliceType).Elem()
	if err := bq.FindHasManyIn(resultSlice.Addr().Interface(), fkColumn, parentKeys); err != nil {
		kklogger.ErrorJ("models:_EagerHasMany#query", err.Error())
	}

	groupByKey := make(map[any]reflect.Value)
	for i := 0; i < resultSlice.Len(); i++ {
		c := resultSlice.Index(i)
		fkVal := c.Elem().FieldByName(fkFieldName)
		if fkVal.Kind() == reflect.Pointer {
			if fkVal.IsNil() {
				continue
			}

			fkVal = fkVal.Elem()
		}

		key := fkVal.Interface()
		if _, exists := groupByKey[key]; !exists {
			groupByKey[key] = reflect.MakeSlice(sliceType, 0, 1)
		}

		groupByKey[key] = reflect.Append(groupByKey[key], c)
	}

	setCacheName := "SetCache" + strings.TrimPrefix(field.Name, "_")
	emptySlice := reflect.MakeSlice(sliceType, 0, 0)
	for key, rvs := range keyToItems {
		group := emptySlice
		if g, ok := groupByKey[key]; ok {
			group = g
		}

		for _, rv := range rvs {
			if m := rv.MethodByName(setCacheName); m.IsValid() {
				m.Call([]reflect.Value{group})
			}
		}
	}
}

// _EagerHasManyComposite batch-loads a has-many association via
// `WHERE (fk1, fk2, ...) IN ((v1, v2, ...), ...)`. Scope conditions (any op) flow through
// as `extra ...LazyCondition` on the querier, not into the tuple. Single-FK fields with a
// scope share this path with a one-column tuple.
func _EagerHasManyComposite(items []reflect.Value, field reflect.StructField, fkFieldNameRaw string, refFieldNameRaw string, childType reflect.Type, scope []LazyCondition) {
	cbq := _GetCompositeBatchQuerierForType(childType)
	if cbq == nil {
		return
	}

	fkNames := _SplitCompositeTag(fkFieldNameRaw)
	refNames := _SplitCompositeTag(refFieldNameRaw)
	if len(refNames) == 0 {
		refNames = []string{"ID"}
	}

	if len(refNames) != len(fkNames) {
		return
	}

	sliceType := reflect.SliceOf(reflect.PointerTo(childType))

	fkColumns := make([]string, len(fkNames))
	for i, fk := range fkNames {
		fkColumns[i] = _ChildColumn(childType, fk)
	}

	keyToItems := make(map[string][]reflect.Value)
	keyToValues := make(map[string][]any)

	for _, rv := range items {
		mv := rv.Elem()
		if !mv.FieldByName(field.Name).IsNil() {
			continue
		}

		values := make([]any, len(refNames))
		keyParts := make([]string, len(refNames))
		skip := false

		for i, ref := range refNames {
			refField := mv.FieldByName(ref)
			if refField.Kind() == reflect.Pointer {
				if refField.IsNil() {
					skip = true
					break
				}
				refField = refField.Elem()
			}
			val := _ExtractUint64(refField)
			if val == 0 {
				skip = true
				break
			}

			values[i] = val
			keyParts[i] = strconv.FormatUint(val, 10)
		}

		if skip {
			continue
		}

		key := strings.Join(keyParts, ":")
		keyToItems[key] = append(keyToItems[key], rv)
		keyToValues[key] = values
	}

	if len(keyToItems) == 0 {
		return
	}

	compositeKeys := make([][]any, 0, len(keyToValues))
	for _, vals := range keyToValues {
		compositeKeys = append(compositeKeys, vals)
	}

	resultSlice := reflect.New(sliceType).Elem()
	if err := cbq.FindHasManyInComposite(resultSlice.Addr().Interface(), fkColumns, compositeKeys, scope...); err != nil {
		kklogger.ErrorJ("models:_EagerHasManyComposite#query", err.Error())
	}

	groupByKey := make(map[string]reflect.Value)
	for i := 0; i < resultSlice.Len(); i++ {
		c := resultSlice.Index(i)
		keyParts := make([]string, len(fkNames))
		for j, fk := range fkNames {
			val := _ExtractUint64(c.Elem().FieldByName(fk))
			keyParts[j] = strconv.FormatUint(val, 10)
		}

		key := strings.Join(keyParts, ":")
		if _, exists := groupByKey[key]; !exists {
			groupByKey[key] = reflect.MakeSlice(sliceType, 0, 1)
		}

		groupByKey[key] = reflect.Append(groupByKey[key], c)
	}

	setCacheName := "SetCache" + strings.TrimPrefix(field.Name, "_")
	emptySlice := reflect.MakeSlice(sliceType, 0, 0)
	for key, rvs := range keyToItems {
		group := emptySlice
		if g, ok := groupByKey[key]; ok {
			group = g
		}

		for _, rv := range rvs {
			if m := rv.MethodByName(setCacheName); m.IsValid() {
				m.Call([]reflect.Value{group})
			}
		}
	}
}

// _ExtractUint64 extracts a uint64 from a reflect.Value.
// Supports model ID types that expose a UInt64() method, and unsigned int kinds.
// Signed int kinds are not supported — v.Uint() panics for them.
func _ExtractUint64(v reflect.Value) uint64 {
	if m := v.MethodByName("UInt64"); m.IsValid() {
		return m.Call(nil)[0].Uint()
	}

	return v.Uint()
}

// _MapKeys returns the keys of a map[uint64]V as a []uint64 slice.
func _MapKeys[V any](m map[uint64]V) []uint64 {
	keys := make([]uint64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// ValidateLazyTags checks all unexported fields with gorm foreignKey tags
// to ensure the referenced FK field exists on the struct.
// Call at init() time to fail fast on misconfigured tags.
func ValidateLazyTags[T any]() {
	t := reflect.TypeFor[T]()
	for field := range t.Fields() {
		gormTag := field.Tag.Get("gorm")
		fk := _ParseGormTagValue(gormTag, "foreignKey")
		if fk == "" {
			continue
		}

		for part := range strings.SplitSeq(fk, ",") {
			part = strings.TrimSpace(part)
			if _, ok := t.FieldByName(part); !ok {
				panic("models: lazy field " + t.Name() + "." + field.Name +
					" references non-existent FK field: " + part)
			}
		}

		ref := _ParseGormTagValue(gormTag, "references")
		if ref != "" {
			for part := range strings.SplitSeq(ref, ",") {
				part = strings.TrimSpace(part)
				if _, ok := t.FieldByName(part); !ok {
					panic("models: lazy field " + t.Name() + "." + field.Name +
						" references non-existent references field: " + part)
				}
			}
		}
	}
}

// ─── Reload ─────────────────────────────────────────────────────────────────

// Reload reloads all fields of model by its primary key (in-place overwrite).
func Reload[T any](model *T) error {
	id := _ExtractPrimaryKey(model)
	if id == nil {
		return fmt.Errorf("model has no primary key value")
	}

	childType := reflect.TypeFor[T]()
	q := _GetQuerierForType(childType)
	if q == nil {
		return fmt.Errorf("no querier registered for %s", childType.Name())
	}

	if err := q.FindById(model, id); err != nil {
		return fmt.Errorf("reload %s(id=%v): %w", childType.Name(), id, err)
	}

	return nil
}

// ReloadAll batch-reloads a slice of partially-loaded models by their primary keys.
func ReloadAll[T any](items []*T) error {
	if len(items) == 0 {
		return nil
	}

	childType := reflect.TypeFor[T]()
	bq := _GetBatchQuerierForType(childType)
	if bq == nil {
		for _, item := range items {
			if err := Reload(item); err != nil {
				return err
			}
		}

		return nil
	}

	ids := _ExtractPrimaryKeys(items)
	reloaded := make([]*T, 0, len(items))
	if err := bq.FindByIds(&reloaded, ids); err != nil {
		return fmt.Errorf("reload all %s: %w", childType.Name(), err)
	}

	_MergeReloadedBack(items, reloaded)
	return nil
}

// _ExtractPrimaryKey extracts the primary key value from model via reflection.
func _ExtractPrimaryKey(model any) any {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	t := val.Type()
	var field reflect.Value

	for i := 0; i < t.NumField(); i++ {
		gormTag := t.Field(i).Tag.Get("gorm")
		if strings.Contains(strings.ToLower(gormTag), "primarykey") {
			field = val.Field(i)
			break
		}
	}

	if !field.IsValid() {
		for _, name := range []string{"ID", "Id", "id"} {
			f := val.FieldByName(name)
			if f.IsValid() {
				field = f
				break
			}
		}
	}

	if !field.IsValid() || field.IsZero() {
		return nil
	}

	return field.Interface()
}

// _ExtractPrimaryKeys returns the primary key values for a slice of models.
func _ExtractPrimaryKeys[T any](items []*T) []any {
	ids := make([]any, 0, len(items))
	for _, item := range items {
		if id := _ExtractPrimaryKey(item); id != nil {
			ids = append(ids, id)
		}
	}

	return ids
}

// _MergeReloadedBack overwrites each item in items with the fully-reloaded counterpart
// matched by primary key.
func _MergeReloadedBack[T any](items []*T, reloaded []*T) {
	if len(reloaded) == 0 {
		return
	}

	byID := make(map[any]*T, len(reloaded))
	for _, r := range reloaded {
		if id := _ExtractPrimaryKey(r); id != nil {
			byID[id] = r
		}
	}

	for _, item := range items {
		id := _ExtractPrimaryKey(item)
		if id == nil {
			continue
		}

		if full, ok := byID[id]; ok {
			*item = *full
		}
	}
}
