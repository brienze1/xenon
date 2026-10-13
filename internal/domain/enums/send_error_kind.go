package enums

import "strings"

type SendErrorKind string

func (s SendErrorKind) Equals(raw string) bool {
	return strings.EqualFold(string(s), raw)
}

func (s SendErrorKind) String() string {
	return string(s)
}
