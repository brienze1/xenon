package enums

import "strings"

type AIActivityOutcome string

func (v AIActivityOutcome) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v AIActivityOutcome) String() string {
	return string(v)
}
