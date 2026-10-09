package enums

import "strings"

type AuthType string

func (a AuthType) Equals(s string) bool {
	return strings.EqualFold(string(a), s)
}

func (a AuthType) String() string {
	return string(a)
}
