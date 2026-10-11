package dto

type CheckScript struct {
	Text string `json:"text"`
}

func (c CheckScript) LogValue() any {
	return map[string]any{"length": len(c.Text)}
}
