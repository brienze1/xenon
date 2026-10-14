package enums

import "strings"

type AIToolEffect string

func (v AIToolEffect) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v AIToolEffect) String() string {
	return string(v)
}
