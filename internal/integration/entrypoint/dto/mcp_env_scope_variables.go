package dto

type MCPEnvScopeVariables struct {
	Scope string            `json:"scope"`
	Vars  []MCPVariableLine `json:"vars"`
}

func NewMCPEnvScopeVariables(table VariableTable) MCPEnvScopeVariables {
	return MCPEnvScopeVariables{Scope: table.Scope, Vars: NewMCPVariableLines(table.Rows)}
}
