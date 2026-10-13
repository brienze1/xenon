package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSandboxScope struct {
	Active *bool                   `json:"active,omitempty"`
	Name   string                  `json:"name,omitempty"`
	Values []ScriptSandboxVariable `json:"values"`
}

func NewScriptSandboxScope(scope entity.ScriptScope) ScriptSandboxScope {
	values := make([]ScriptSandboxVariable, 0, len(scope.Values))
	for _, variable := range scope.Values {
		values = append(values, NewScriptSandboxVariable(variable))
	}
	return ScriptSandboxScope{Name: scope.Name, Values: values}
}

func (s ScriptSandboxScope) ToEntity(started entity.ScriptScope) entity.ScriptScope {
	values := make([]entity.ScriptVariable, 0, len(s.Values))
	for _, variable := range s.Values {
		values = append(values, variable.ToEntity())
	}
	return entity.ScriptScope{Name: started.Name, Values: values, Inactive: started.Inactive}
}
