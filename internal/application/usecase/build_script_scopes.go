package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

// BuildScriptScopes gives scripts the values of each scope in file order; secrets without a value or in a locked vault are left out.
func BuildScriptScopes(environment *entity.Environment, ancestors []entity.Collection, globals []entity.VariableRow, local entity.LocalVariables, vault entity.Vault) entity.ScriptScopes {
	scopes := entity.ScriptScopes{
		Environment: entity.NewInactiveScriptScope(),
		Collection:  entity.NewScriptScope("", nil),
		Globals:     scriptScope(variablescope.Globals, "", entity.NewGlobalsScopeRef(), globals, local, vault),
		Local:       entity.NewScriptScope("", nil),
	}
	if environment != nil {
		scopes.Environment = scriptScope(variablescope.Environment, environment.Name, environment.ScopeRef(), environment.Variables, local, vault)
	}
	if collection, found := entity.RootCollection(ancestors); found {
		scopes.Collection = scriptScope(variablescope.Collection, collection.Name, collection.ScopeRef(), collection.OwnVariables(), local, vault)
	}
	return scopes
}

// BuildScriptVault is what pm.vault reads: the workspace globals' secret rows that have a value on this machine.
func BuildScriptVault(globals []entity.VariableRow, vault entity.Vault) map[string]string {
	values := map[string]string{}
	if vault.IsLocked() {
		return values
	}
	secrets := vault.SecretsFor(entity.NewGlobalsScopeRef())
	for _, row := range globals {
		if value, found := secrets[row.Key]; found && row.Secret && row.IsActive() {
			values[row.Key] = value
		}
	}
	return values
}

func scriptScope(kind enums.VariableScope, name string, scope entity.ScopeRef, rows []entity.VariableRow, local entity.LocalVariables, vault entity.Vault) entity.ScriptScope {
	values := make([]entity.ScriptVariable, 0, len(rows))
	current := local.CurrentValues(scope)
	secrets := vault.SecretsFor(scope)
	for _, row := range rows {
		if !row.IsActive() {
			continue
		}
		variable := row.ContextVariable(kind, current, secrets)
		if !variable.IsResolvable() || (row.Secret && vault.IsLocked()) {
			continue
		}
		scriptVariable := entity.NewScriptVariable(row.Key, variable.Value)
		scriptVariable.Secret = row.Secret
		values = append(values, scriptVariable)
	}
	return entity.NewScriptScope(name, values)
}
