package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvReference struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Value string `json:"value"`
}

// NewMCPEnvReference masks the value of a secret variable.
func NewMCPEnvReference(usage VariableUsage) MCPEnvReference {
	reference := MCPEnvReference{Name: usage.Name, Scope: usage.Scope, Value: usage.DisplayValue}
	if usage.Secret {
		reference.Value = entity.MaskedValue
	}
	return reference
}
