package enums

import "strings"

type RunStatus string

func (v RunStatus) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v RunStatus) String() string {
	return string(v)
}
