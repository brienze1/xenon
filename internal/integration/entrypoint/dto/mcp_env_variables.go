package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvVariables struct {
	Scope string           `json:"scope"`
	Vars  []MCPVariableRow `json:"vars"`
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
