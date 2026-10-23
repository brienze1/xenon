package enums

import "strings"

type RunRowOutcome string

func (v RunRowOutcome) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v RunRowOutcome) String() string {
	return string(v)
}
