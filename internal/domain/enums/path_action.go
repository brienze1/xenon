package enums

import "strings"

type PathAction string

func (p PathAction) Equals(s string) bool {
	return strings.EqualFold(string(p), s)
}

func (p PathAction) String() string {
	return string(p)
}
