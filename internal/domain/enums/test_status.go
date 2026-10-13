package enums

import "strings"

type TestStatus string

func (v TestStatus) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v TestStatus) String() string {
	return string(v)
}
