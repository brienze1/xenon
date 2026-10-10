package entity

import "slices"

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

type VariableInspection struct {
	References []ResolvedReference
	Variables  []VariableUsage
}

func NewVariableInspection() VariableInspection {
	return VariableInspection{References: []ResolvedReference{}, Variables: []VariableUsage{}}
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
