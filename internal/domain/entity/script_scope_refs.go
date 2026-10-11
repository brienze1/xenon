package entity

import (
	"maps"
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

// ScriptScopeRefs are the scopes script writes go to: the active environment, the root collection and the globals.
func ScriptScopeRefs(environment *Environment, ancestors []Collection) map[enums.VariableScope]ScopeRef {
	refs := map[enums.VariableScope]ScopeRef{variablescope.Globals: NewGlobalsScopeRef()}
	if environment != nil {
		refs[variablescope.Environment] = environment.ScopeRef()
	}
	if collection, found := RootCollection(ancestors); found {
		refs[variablescope.Collection] = collection.ScopeRef()
	}
	return refs
}

func AppendScopeRef(refs []ScopeRef, ref ScopeRef) []ScopeRef {
	if slices.Contains(refs, ref) {
		return refs
	}
	return append(refs, ref)
}

// VaultChangesFrom lists, in key order, the vault values a script set or changed and the keys it removed.
func VaultChangesFrom(before, after map[string]string) ([]KeyValue, []string) {
	set := make([]KeyValue, 0)
	for _, key := range slices.Sorted(maps.Keys(after)) {
		if old, found := before[key]; !found || old != after[key] {
			set = append(set, NewKeyValue(key, after[key]))
		}
	}
	unset := make([]string, 0)
	for _, key := range slices.Sorted(maps.Keys(before)) {
		if _, kept := after[key]; !kept {
			unset = append(unset, key)
		}
	}
	return set, unset
}
