package enums

import "strings"

type TokenPlacement string

func (t TokenPlacement) Equals(s string) bool {
	return strings.EqualFold(string(t), s)
}

func (t TokenPlacement) String() string {
	return string(t)
}
