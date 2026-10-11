package enums

import "strings"

type ScriptMemberKind string

func (v ScriptMemberKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ScriptMemberKind) String() string {
	return string(v)
}
