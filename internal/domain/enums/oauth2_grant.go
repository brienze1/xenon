package enums

import "strings"

type OAuth2Grant string

func (o OAuth2Grant) Equals(s string) bool {
	return strings.EqualFold(string(o), s)
}

func (o OAuth2Grant) String() string {
	return string(o)
}
