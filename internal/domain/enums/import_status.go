package enums

import "strings"

type ImportStatus string

func (i ImportStatus) Equals(s string) bool {
	return strings.EqualFold(string(i), s)
}

func (i ImportStatus) String() string {
	return string(i)
}
