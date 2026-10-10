package enums

import "strings"

type ClashResolution string

func (c ClashResolution) Equals(s string) bool {
	return strings.EqualFold(string(c), s)
}

func (c ClashResolution) String() string {
	return string(c)
}
