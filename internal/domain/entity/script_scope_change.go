package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

var PersistedScriptScopes = []enums.VariableScope{variablescope.Environment, variablescope.Collection, variablescope.Globals}

// ScriptScopeChange is what a script did to one saved scope; Secret marks changes to the vault.
type ScriptScopeChange struct {
	Ref    ScopeRef
	Set    []KeyValue
	Unset  []string
	Secret bool
}

func (c ScriptScopeChange) IsEmpty() bool {
	return len(c.Set) == 0 && len(c.Unset) == 0
}

// ChangesFrom lists, in save order, the persisted scopes with a reference and then the vault that the script changed.
func (s ScriptScopes) ChangesFrom(before ScriptScopes, refs map[enums.VariableScope]ScopeRef, vaultBefore, vaultAfter map[string]string) []ScriptScopeChange {
	changes := make([]ScriptScopeChange, 0, len(PersistedScriptScopes)+1)
	for _, kind := range PersistedScriptScopes {
		ref, found := refs[kind]
		set, unset := s.Scope(kind).ChangesFrom(before.Scope(kind))
		if change := (ScriptScopeChange{Ref: ref, Set: set, Unset: unset}); found && !change.IsEmpty() {
			changes = append(changes, change)
		}
	}
	set, unset := VaultChangesFrom(vaultBefore, vaultAfter)
	if vault := (ScriptScopeChange{Ref: NewGlobalsScopeRef(), Set: set, Unset: unset, Secret: true}); !vault.IsEmpty() {
		changes = append(changes, vault)
	}
	return changes
}
