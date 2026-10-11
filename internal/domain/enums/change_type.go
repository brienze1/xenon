package enums

import "strings"

type ChangeType string

func (c ChangeType) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ChangeType) String() string {
	return string(c)
}
