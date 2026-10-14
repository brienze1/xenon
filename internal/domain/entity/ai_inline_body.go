package entity

const AIInlineBodyBytes = 16384

type AIInlineBody struct {
	Text      string
	Truncated bool
	Size      int
}

// NewAIInlineBody caps an already masked body at 16 KB on a rune boundary and keeps its full size.
func NewAIInlineBody(masked string) AIInlineBody {
	if len(masked) <= AIInlineBodyBytes {
		return AIInlineBody{Text: masked, Size: len(masked)}
	}
	return AIInlineBody{Text: masked[:runeBoundary(masked, AIInlineBodyBytes)], Truncated: true, Size: len(masked)}
}
