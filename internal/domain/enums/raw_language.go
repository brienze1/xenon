package enums

import "strings"

type RawLanguage string

func (r RawLanguage) Equals(s string) bool {
	return strings.EqualFold(string(r), s)
}

func (r RawLanguage) String() string {
	return string(r)
}
