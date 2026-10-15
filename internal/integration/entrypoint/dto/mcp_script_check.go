package dto

type MCPScriptCheck struct {
	Text string `json:"text" description:"Script source to check."`
}

func (m *MCPScriptCheck) CheckScript() CheckScript {
	return NewCheckScript(m.Text)
}

func (m MCPScriptCheck) LogValue() any {
	return map[string]any{"length": len(m.Text)}
}
