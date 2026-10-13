package enums

import "strings"

type SendOutcome string

func (v SendOutcome) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v SendOutcome) String() string {
	return string(v)
}
