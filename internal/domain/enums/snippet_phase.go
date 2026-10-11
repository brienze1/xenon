package enums

import "strings"

type SnippetPhase string

func (v SnippetPhase) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v SnippetPhase) String() string {
	return string(v)
}
