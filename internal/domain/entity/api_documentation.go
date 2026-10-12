package entity

// ApiDocumentation is a collection's or folder's OpenAPI document as text and as the view rendered from that text.
type ApiDocumentation struct {
	Text     string
	Document ApiDocument
}

func (a ApiDocumentation) LogValue() any {
	return map[string]any{"title": a.Document.Title, "operations": len(a.Document.Operations), "bytes": len(a.Text)}
}
