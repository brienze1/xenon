package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type VariableUsage struct {
	UsedIn   []string
	Resolved ResolvedReference
}

func (v VariableUsage) withField(field string) VariableUsage {
	if !slices.Contains(v.UsedIn, field) {
		v.UsedIn = append(append([]string{}, v.UsedIn...), field)
	}
	return v
}

// InspectedScope is a variable table the request can use, strongest first.
type InspectedScope struct {
	Scope enums.VariableScope
	Ref   ScopeRef
}

type VariableInspection struct {
	References []ResolvedReference
	Variables  []VariableUsage
	Scopes     []InspectedScope
}

func NewVariableInspection(context VariableContext) VariableInspection {
	scopes := make([]InspectedScope, 0, len(context.Layers))
	for _, layer := range context.Layers {
		scopes = append(scopes, InspectedScope{Scope: layer.Scope, Ref: layer.Ref})
	}
	return VariableInspection{References: []ResolvedReference{}, Variables: []VariableUsage{}, Scopes: scopes}
}

func (v VariableInspection) WithReference(reference ResolvedReference) VariableInspection {
	v.References = append(v.References, reference)
	return v
}

// WithUsage records that a field uses the variable, once per name and field.
func (v VariableInspection) WithUsage(resolved ResolvedReference, field string) VariableInspection {
	for index, usage := range v.Variables {
		if usage.Resolved.Reference.Name == resolved.Reference.Name {
			v.Variables[index] = usage.withField(field)
			return v
		}
	}
	v.Variables = append(v.Variables, VariableUsage{Resolved: resolved, UsedIn: []string{field}})
	return v
}

func (v VariableInspection) LogValue() any {
	return map[string]any{"references": len(v.References), "variables": len(v.Variables)}
}
