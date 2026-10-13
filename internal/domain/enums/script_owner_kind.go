package enums

import "strings"

type ScriptOwnerKind string

func (v ScriptOwnerKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ScriptOwnerKind) String() string {
	return string(v)
}
