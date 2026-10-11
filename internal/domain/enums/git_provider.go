package enums

import "strings"

type GitProvider string

func (g GitProvider) Equals(s string) bool {
	return strings.EqualFold(string(g), s)
}

func (g GitProvider) String() string {
	return string(g)
}
