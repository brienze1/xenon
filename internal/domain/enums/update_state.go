package enums

import "strings"

type UpdateState string

func (v UpdateState) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v UpdateState) String() string {
	return string(v)
}
