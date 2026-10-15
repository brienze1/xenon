package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvVariable struct {
	Scope       string `json:"scope" description:"Variable scope: globals, env:<id> or collection:<id>." required_for:"add_var,set_current,reset,persist"`
	Key         string `json:"key" description:"Variable name." required_for:"add_var,set_current,reset,persist"`
	Value       string `json:"value" description:"Value to set; write-only for secrets." used_by:"add_var,set_current"`
	Description string `json:"description" description:"Variable description." used_by:"add_var"`
	Secret      bool   `json:"secret" description:"Store the variable as a secret." used_by:"add_var"`
}

func (m *MCPEnvVariable) AddVariable(localId string) AddVariable {
	return NewAddVariable(localId, m.Scope, m.Key, m.Value, m.Description, m.Secret)
}

func (m *MCPEnvVariable) SetCurrentValue(localId string) SetCurrentValue {
	return NewSetCurrentValue(localId, m.Scope, m.Key, m.Value)
}

func (m *MCPEnvVariable) VariableKeys(localId string) VariableKeys {
	return NewVariableKeys(localId, m.Scope, []string{m.Key})
}

func (m *MCPEnvVariable) Target() entity.AIActivityTarget {
	return entity.AIActivityTarget{Name: m.Key, Path: m.Scope}
}
