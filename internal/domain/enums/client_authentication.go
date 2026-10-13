package enums

import "strings"

type ClientAuthentication string

func (c ClientAuthentication) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ClientAuthentication) String() string {
	return string(c)
}
