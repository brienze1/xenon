package dto

type AIGuarded struct {
	Text string `json:"text"`
}

func NewAIGuarded(text string) AIGuarded {
	return AIGuarded{Text: text}
}

func (a AIGuarded) LogValue() any {
	return map[string]any{"length": len(a.Text)}
}
