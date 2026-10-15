package enums

import "strings"

type UpdateFailure string

func (v UpdateFailure) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v UpdateFailure) String() string {
	return string(v)
}
