package enums

import "strings"

type ImportItemKind string

func (i ImportItemKind) Equals(s string) bool {
	return strings.EqualFold(string(i), s)
}

func (i ImportItemKind) String() string {
	return string(i)
}
