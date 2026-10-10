package enums

import "strings"

type ImportOrigin string

func (i ImportOrigin) Equals(s string) bool {
	return strings.EqualFold(string(i), s)
}

func (i ImportOrigin) String() string {
	return string(i)
}
