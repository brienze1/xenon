package enums

import "strings"

type ClashKind string

func (c ClashKind) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ClashKind) String() string {
	return string(c)
}
