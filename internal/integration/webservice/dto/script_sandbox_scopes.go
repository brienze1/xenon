package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// ScriptSandboxScopes uses the property names postman-sandbox gives the scopes.
type ScriptSandboxScopes struct {
	Environment         ScriptSandboxScope `json:"environment"`
	CollectionVariables ScriptSandboxScope `json:"collectionVariables"`
	Globals             ScriptSandboxScope `json:"globals"`
	Local               ScriptSandboxScope `json:"local"`
}

func NewScriptSandboxScopes(scopes entity.ScriptScopes) ScriptSandboxScopes {
	environment := NewScriptSandboxScope(scopes.Environment)
	if scopes.Environment.Inactive {
		active := false
		environment.Active = &active
	}
	return ScriptSandboxScopes{
		Environment:         environment,
		CollectionVariables: NewScriptSandboxScope(scopes.Collection),
		Globals:             NewScriptSandboxScope(scopes.Globals),
		Local:               NewScriptSandboxScope(scopes.Local),
	}
}

// ToEntity keeps what the script cannot change about a scope (its name, whether it is active) from the scopes it started with.
func (s ScriptSandboxScopes) ToEntity(started entity.ScriptScopes) entity.ScriptScopes {
	return entity.ScriptScopes{
		Environment: s.Environment.ToEntity(started.Environment),
		Collection:  s.CollectionVariables.ToEntity(started.Collection),
		Globals:     s.Globals.ToEntity(started.Globals),
		Local:       s.Local.ToEntity(started.Local),
	}
}
