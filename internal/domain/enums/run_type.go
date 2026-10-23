package enums

import "strings"

type RunType string

func (v RunType) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v RunType) String() string {
	return string(v)
}
