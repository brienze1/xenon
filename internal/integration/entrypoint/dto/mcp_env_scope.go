package dto

type MCPEnvScope struct {
	Scope string `json:"scope"`
}

func (m *MCPEnvScope) VariableTableReference(localId string) VariableTableReference {
	return NewScopeVariableTableReference(localId, m.Scope)
}
