package dto

type AIGuard struct {
	LocalId string `json:"local_id"`
	Text    string `json:"text"`
}

type AIGuarded struct {
	Text string `json:"text"`
}

func NewAIGuarded(text string) AIGuarded {
	return AIGuarded{Text: text}
}

func (a AIGuard) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "length": len(a.Text)}
}

func (a AIGuarded) LogValue() any {
	return map[string]any{"length": len(a.Text)}
}
