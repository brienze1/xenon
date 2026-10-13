package enums

import "strings"

type AIChangeKind string

func (v AIChangeKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v AIChangeKind) String() string {
	return string(v)
}
