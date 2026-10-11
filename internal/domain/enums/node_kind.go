package enums

import "strings"

type NodeKind string

func (n NodeKind) Equals(s string) bool {
	return strings.EqualFold(string(n), s)
}

func (n NodeKind) String() string {
	return string(n)
}
