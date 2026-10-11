package enums

import "strings"

type SignInStatus string

func (v SignInStatus) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v SignInStatus) String() string {
	return string(v)
}
