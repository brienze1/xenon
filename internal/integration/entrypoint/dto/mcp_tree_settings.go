package dto

type MCPTreeSettings struct {
	Scripts     *Scripts              `json:"scripts,omitempty"`
	Auth        Auth                  `json:"auth"`
	Description string                `json:"description,omitempty"`
	Headers     []KeyValue            `json:"headers"`
	Variables   []MCPTreeVariableLine `json:"variables,omitempty"`
}

func NewMCPTreeSettings(settings ItemSettings) MCPTreeSettings {
	return MCPTreeSettings{Auth: settings.Auth, Headers: settings.Headers, Scripts: settings.Scripts, Description: settings.Description}
}

func (m MCPTreeSettings) WithVariables(table VariableTable) MCPTreeSettings {
	m.Variables = NewMCPTreeVariableLines(table.Rows)
	return m
}
