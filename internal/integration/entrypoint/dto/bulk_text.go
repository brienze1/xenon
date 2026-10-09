package dto

type BulkText struct {
	Text string `json:"text"`
}

func NewBulkText(text string) BulkText {
	return BulkText{Text: text}
}

func (b BulkText) LogValue() any {
	return map[string]any{"length": len(b.Text)}
}
