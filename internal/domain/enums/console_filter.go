package enums

import "strings"

type ConsoleFilter string

func (v ConsoleFilter) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ConsoleFilter) String() string {
	return string(v)
}
