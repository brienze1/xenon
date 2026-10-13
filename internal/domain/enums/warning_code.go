package enums

import "strings"

type WarningCode string

func (w WarningCode) Equals(s string) bool {
	return strings.EqualFold(string(w), s)
}

func (w WarningCode) String() string {
	return string(w)
}
