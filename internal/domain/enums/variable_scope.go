package enums

import "strings"

type VariableScope string

func (v VariableScope) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v VariableScope) String() string {
	return string(v)
}
