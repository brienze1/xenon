package entity

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

func (v VariableInspection) WithNestedUsages(resolved ResolvedReference, field string, context VariableContext, dynamicNames []string) VariableInspection {
	visited := map[string]bool{}
	pending := []ResolvedReference{resolved}
	for len(pending) > 0 {
		current := pending[0]
		pending = pending[1:]
		if visited[current.Reference.Name] {
			continue
		}
		visited[current.Reference.Name] = true
		v = v.WithUsage(current, field)
		variable, found := context.Lookup(current.Reference.Name)
		if !found {
			continue
		}
		for _, nested := range TokenizeReferences(field, variable.Value) {
			pending = append(pending, context.Inspect(nested, dynamicNames))
		}
	}
	return v
}

func (v VariableInspection) LogValue() any {
	return map[string]any{"references": len(v.References), "variables": len(v.Variables)}
}
