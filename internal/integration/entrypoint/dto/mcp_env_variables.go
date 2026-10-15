package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvVariables struct {
	Scope string           `json:"scope" description:"Variable scope: globals, env:<id> or collection:<id>." required_for:"set_vars"`
	Vars  []MCPVariableRow `json:"vars" description:"The full variable table, replacing the current one; secret_value only on secret rows."`
}

func (m *MCPEnvVariables) Target() entity.AIActivityTarget {
	return entity.AIActivityTarget{Path: m.Scope}
}

func (m *MCPEnvVariables) VariableScope() string {
	return m.Scope
}

func (m *MCPEnvVariables) VariableRows() []MCPVariableRow {
	return m.Vars
}
