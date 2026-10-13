package enums

import "strings"

type ScriptPhase string

func (p ScriptPhase) Equals(s string) bool {
	return strings.EqualFold(string(p), s)
}

func (p ScriptPhase) String() string {
	return string(p)
}
