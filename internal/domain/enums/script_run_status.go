package enums

import "strings"

type ScriptRunStatus string

func (v ScriptRunStatus) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ScriptRunStatus) String() string {
	return string(v)
}
