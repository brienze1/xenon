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
