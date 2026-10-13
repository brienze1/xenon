package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Documentation struct {
	Yaml     string      `json:"yaml"`
	Document ApiDocument `json:"document"`
}

func NewDocumentation(documentation entity.ApiDocumentation) Documentation {
	return Documentation{Yaml: documentation.Text, Document: NewApiDocument(documentation.Document)}
}

func (d Documentation) LogValue() any {
	return map[string]any{"title": d.Document.Title, "operations": len(d.Document.Operations), "bytes": len(d.Yaml)}
}
