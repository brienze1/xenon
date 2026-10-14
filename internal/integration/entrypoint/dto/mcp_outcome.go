package dto

type MCPOutcome struct {
	Text   string `json:"text"`
	Failed bool   `json:"failed"`
}

func NewMCPOutcome(text string, failed bool) MCPOutcome {
	return MCPOutcome{Text: text, Failed: failed}
}

func (m MCPOutcome) LogValue() any {
	return map[string]any{"length": len(m.Text), "failed": m.Failed}
}
