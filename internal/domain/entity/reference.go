package entity

import "strings"

const dynamicVariablePrefix = "$"

type Reference struct {
	Name  string
	Field string
	Start int
	End   int
}

func (r Reference) Placeholder() string {
	return "{{" + r.Name + "}}"
}

func (r Reference) IsDynamic() bool {
	return strings.HasPrefix(r.Name, dynamicVariablePrefix)
}
