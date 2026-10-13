package enums

import "strings"

type ImportKind string

func (i ImportKind) Equals(s string) bool {
	return strings.EqualFold(string(i), s)
}

func (i ImportKind) String() string {
	return string(i)
}
