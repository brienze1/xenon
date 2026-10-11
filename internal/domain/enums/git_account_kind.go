package enums

import "strings"

type GitAccountKind string

func (g GitAccountKind) Equals(s string) bool {
	return strings.EqualFold(string(g), s)
}

func (g GitAccountKind) String() string {
	return string(g)
}
