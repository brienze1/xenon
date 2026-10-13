package enums

import "strings"

type ScriptErrorKind string

func (v ScriptErrorKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ScriptErrorKind) String() string {
	return string(v)
}
