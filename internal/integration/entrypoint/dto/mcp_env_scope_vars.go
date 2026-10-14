package dto

type MCPEnvScopeVars struct {
	Name string            `json:"name"`
	Vars []MCPVariableLine `json:"vars"`
}

func NewMCPEnvScopeVars(name string, table VariableTable) MCPEnvScopeVars {
	return MCPEnvScopeVars{Name: name, Vars: NewMCPVariableLines(table.Rows)}
}
