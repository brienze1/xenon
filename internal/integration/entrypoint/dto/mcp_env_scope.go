package dto

type MCPEnvScope struct {
	Scope string `json:"scope" description:"Variable scope: globals, env:<id> or collection:<id>." required_for:"get"`
}

func (m *MCPEnvScope) VariableTableReference(localId string) VariableTableReference {
	return NewScopeVariableTableReference(localId, m.Scope)
}
