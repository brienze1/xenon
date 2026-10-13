package enums

import "strings"

type ReferenceStatus string

func (r ReferenceStatus) Equals(s string) bool {
	return strings.EqualFold(string(r), s)
}

func (r ReferenceStatus) String() string {
	return string(r)
}
