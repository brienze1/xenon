package enums

import "strings"

type LinkMode string

func (l LinkMode) Equals(s string) bool {
	return strings.EqualFold(string(l), s)
}

func (l LinkMode) String() string {
	return string(l)
}
