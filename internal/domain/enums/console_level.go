package enums

import "strings"

type ConsoleLevel string

func (v ConsoleLevel) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ConsoleLevel) String() string {
	return string(v)
}
