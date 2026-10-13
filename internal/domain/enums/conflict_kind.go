package enums

import "strings"

type ConflictKind string

func (c ConflictKind) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ConflictKind) String() string {
	return string(c)
}
