package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableInspection struct {
	References []VariableReference `json:"references"`
	Variables  []VariableUsage     `json:"variables"`
	Scopes     []InspectedScope    `json:"scopes"`
}

func NewVariableInspection(inspection entity.VariableInspection) VariableInspection {
	result := VariableInspection{
		References: make([]VariableReference, 0, len(inspection.References)),
		Variables:  make([]VariableUsage, 0, len(inspection.Variables)),
		Scopes:     make([]InspectedScope, 0, len(inspection.Scopes)),
	}
	for _, scope := range inspection.Scopes {
		result.Scopes = append(result.Scopes, NewInspectedScope(scope))
	}
	for _, reference := range inspection.References {
		result.References = append(result.References, NewVariableReference(reference))
	}
	for _, usage := range inspection.Variables {
		result.Variables = append(result.Variables, NewVariableUsage(usage))
	}
	return result
}

func (v VariableInspection) LogValue() any {
	return map[string]any{"references": len(v.References), "variables": len(v.Variables)}
}
