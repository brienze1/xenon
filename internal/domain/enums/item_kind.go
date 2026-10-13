package enums

import "strings"

type ItemKind string

func (i ItemKind) Equals(s string) bool {
	return strings.EqualFold(string(i), s)
}

func (i ItemKind) String() string {
	return string(i)
}
