package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableReference struct {
	Field        string `json:"field"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Scope        string `json:"scope"`
	DisplayValue string `json:"display_value"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
	Secret       bool   `json:"secret"`
}

func NewVariableReference(resolved entity.ResolvedReference) VariableReference {
	return VariableReference{
		Field:        resolved.Reference.Field,
		Name:         resolved.Reference.Name,
		Status:       resolved.Status.String(),
		Scope:        resolved.Scope.String(),
		DisplayValue: resolved.DisplayValue(),
		Start:        resolved.Reference.Start,
		End:          resolved.Reference.End,
		Secret:       resolved.Secret,
	}
}
