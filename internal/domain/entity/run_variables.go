package entity

import (
	"maps"
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

type RunVariables struct {
	Values       map[enums.VariableScope][]KeyValue
	Removed      map[enums.VariableScope][]string
	Vault        map[string]string
	VaultRemoved []string
}

func NewRunVariables() *RunVariables {
	return &RunVariables{Values: map[enums.VariableScope][]KeyValue{}, Removed: map[enums.VariableScope][]string{}, Vault: map[string]string{}, VaultRemoved: []string{}}
}

func (r *RunVariables) Apply(before, after ScriptScopes, vaultBefore, vaultAfter map[string]string) {
	for _, kind := range PersistedScriptScopes {
		set, unset := after.Scope(kind).ChangesFrom(before.Scope(kind))
		for _, value := range set {
			r.set(kind, value)
		}
		for _, name := range unset {
			r.unset(kind, name)
		}
	}
	set, unset := VaultChangesFrom(vaultBefore, vaultAfter)
	for _, value := range set {
		r.Vault[value.Key] = value.Value
		r.VaultRemoved = slices.DeleteFunc(r.VaultRemoved, func(name string) bool { return name == value.Key })
	}
	for _, name := range unset {
		delete(r.Vault, name)
		if !slices.Contains(r.VaultRemoved, name) {
			r.VaultRemoved = append(r.VaultRemoved, name)
		}
	}
}

func (r *RunVariables) Over(scopes ScriptScopes) ScriptScopes {
	scopes.Environment = r.overScope(variablescope.Environment, scopes.Environment)
	scopes.Collection = r.overScope(variablescope.Collection, scopes.Collection)
	scopes.Globals = r.overScope(variablescope.Globals, scopes.Globals)
	return scopes
}

func (r *RunVariables) OverVault(vault map[string]string) map[string]string {
	updated := maps.Clone(vault)
	if updated == nil {
		updated = map[string]string{}
	}
	for _, name := range r.VaultRemoved {
		delete(updated, name)
	}
	maps.Copy(updated, r.Vault)
	return updated
}

func (r *RunVariables) overScope(kind enums.VariableScope, scope ScriptScope) ScriptScope {
	if scope.Inactive {
		return scope
	}
	for _, name := range r.Removed[kind] {
		scope = scope.WithoutValue(name)
	}
	for _, value := range r.Values[kind] {
		scope = scope.WithValue(value.Key, value.Value)
	}
	return scope
}

func (r *RunVariables) set(kind enums.VariableScope, value KeyValue) {
	r.Removed[kind] = slices.DeleteFunc(r.Removed[kind], func(name string) bool { return name == value.Key })
	values := slices.DeleteFunc(slices.Clone(r.Values[kind]), func(existing KeyValue) bool { return existing.Key == value.Key })
	r.Values[kind] = append(values, value)
}

func (r *RunVariables) unset(kind enums.VariableScope, name string) {
	r.Values[kind] = slices.DeleteFunc(slices.Clone(r.Values[kind]), func(existing KeyValue) bool { return existing.Key == name })
	if !slices.Contains(r.Removed[kind], name) {
		r.Removed[kind] = append(r.Removed[kind], name)
	}
}
