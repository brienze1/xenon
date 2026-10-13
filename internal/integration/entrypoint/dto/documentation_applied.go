package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type DocumentationApplied struct {
	Documentation
	DocumentationPlan
}

func NewDocumentationApplied(result entity.DocumentationResult) DocumentationApplied {
	return DocumentationApplied{DocumentationPlan: NewDocumentationPlan(result.Plan), Documentation: NewDocumentation(result.Documentation)}
}

func (d DocumentationApplied) LogValue() any {
	return map[string]any{"plan": d.DocumentationPlan.LogValue(), "documentation": d.Documentation.LogValue()}
}
