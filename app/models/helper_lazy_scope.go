package models

import (
	"sync"
)

// LazyCondition is one WHERE term applied to every lazy/eager load of a field
// tagged `scope:"<name>"`. The Op vocabulary mirrors app/components/queryfilter
// intentionally — same set of strings, no import cycle.
type LazyCondition struct {
	Column string
	Op     string
	Value  any
}

// LazyOp constants mirror queryfilter.Op. New operators must be added in
// BOTH places (or this set kept narrower than queryfilter's).
const (
	LazyOpEq       = "eq"
	LazyOpNe       = "ne"
	LazyOpGt       = "gt"
	LazyOpGte      = "gte"
	LazyOpLt       = "lt"
	LazyOpLte      = "lte"
	LazyOpLike     = "like"
	LazyOpPrefix   = "prefix"
	LazyOpSuffix   = "suffix"
	LazyOpEmpty    = "empty"
	LazyOpNotEmpty = "notempty"
)

// LazyScope provides extra conditions ANDed into every lazy/eager load of a
// field tagged `scope:"<name>"`. Returning an empty slice skips the filter.
type LazyScope func() []LazyCondition

var _LazyScopeRegistry sync.Map

// RegisterLazyScope binds a scope name to its condition provider. Calls with
// an empty name or nil provider are dropped silently so init order stays loose.
//
// The scaffold ships no scope registrations; downstream projects register their
// own (e.g. a tenant/channel filter) and tag the relevant model fields with
// `scope:"<name>"`. Until a field carries a scope tag this whole path is dormant.
func RegisterLazyScope(name string, scope LazyScope) {
	if name == "" || scope == nil {
		return
	}

	_LazyScopeRegistry.Store(name, scope)
}

// _ResolveLazyScope returns the conditions for the given scope name; nil when
// the name is empty, unregistered, or the provider yields no conditions.
func _ResolveLazyScope(name string) []LazyCondition {
	if name == "" {
		return nil
	}

	v, ok := _LazyScopeRegistry.Load(name)
	if !ok {
		return nil
	}

	conds := v.(LazyScope)()
	if len(conds) == 0 {
		return nil
	}

	return conds
}
