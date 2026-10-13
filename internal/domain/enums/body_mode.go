package enums

import "strings"

type BodyMode string

func (b BodyMode) Equals(s string) bool {
	return strings.EqualFold(string(b), s)
}

func (b BodyMode) String() string {
	return string(b)
}
