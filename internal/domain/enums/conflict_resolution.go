package enums

import "strings"

type ConflictResolution string

func (c ConflictResolution) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ConflictResolution) String() string {
	return string(c)
}
