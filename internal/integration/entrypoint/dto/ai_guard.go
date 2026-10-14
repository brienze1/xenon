package dto

type AIGuard struct {
	LocalId string `json:"local_id"`
	Text    string `json:"text"`
}

func NewAIGuard(localId, text string) AIGuard {
	return AIGuard{LocalId: localId, Text: text}
}

func (a AIGuard) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "length": len(a.Text)}
}

func (a *AIGuard) WorkspaceId() string {
	return a.LocalId
}

func (a *AIGuard) GuardedText() string {
	return a.Text
}
