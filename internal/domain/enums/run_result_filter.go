package enums

import "strings"

type RunResultFilter string

func (v RunResultFilter) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v RunResultFilter) String() string {
	return string(v)
}
