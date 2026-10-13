package enums

import "strings"

type AIAccessState string

func (v AIAccessState) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v AIAccessState) String() string {
	return string(v)
}
