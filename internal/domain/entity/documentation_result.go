package entity

// DocumentationResult is a saved documentation edit: what it changed and the collection's document afterwards.
type DocumentationResult struct {
	Documentation ApiDocumentation
	Plan          DocumentationPlan
}

func (d DocumentationResult) LogValue() any {
	return map[string]any{"plan": d.Plan.LogValue(), "documentation": d.Documentation.LogValue()}
}
