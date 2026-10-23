package enums

import "strings"

type LoadProfileKind string

func (v LoadProfileKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v LoadProfileKind) String() string {
	return string(v)
}
