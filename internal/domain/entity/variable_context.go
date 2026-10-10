package entity

import "strings"

type VariableContext struct {
	Layers []VariableLayer
}

func NewVariableContext(layers ...VariableLayer) VariableContext {
	return VariableContext{Layers: append([]VariableLayer{}, layers...)}
}

func (v VariableContext) Lookup(name string) (ContextVariable, bool) {
	for _, layer := range v.Layers {
		if variable, found := layer.Lookup(name); found {
			return variable, true
		}
	}
	return ContextVariable{}, false
}

func (v VariableContext) SecretValues() []string {
	values := make([]string, 0)
	for _, layer := range v.Layers {
		values = append(values, layer.secretValues()...)
	}
	return values
}

// Masked replaces every secret value of the context found in the text with the mask.
func (v VariableContext) Masked(text string) string {
	for _, secret := range v.SecretValues() {
		text = strings.ReplaceAll(text, secret, MaskedValue)
	}
	return text
}

// Variables lists every variable once, strongest scope first (a name defined in several scopes keeps the strongest).
func (v VariableContext) Variables() []ContextVariable {
	seen := make(map[string]bool)
	variables := make([]ContextVariable, 0)
	for _, layer := range v.Layers {
		for _, variable := range layer.sortedVariables() {
			if !seen[variable.Name] {
				seen[variable.Name] = true
				variables = append(variables, variable)
			}
		}
	}
	return variables
}
