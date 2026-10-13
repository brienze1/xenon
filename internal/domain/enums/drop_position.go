package enums

import "strings"

type DropPosition string

func (d DropPosition) Equals(s string) bool {
	return strings.EqualFold(string(d), s)
}

func (d DropPosition) String() string {
	return string(d)
}
