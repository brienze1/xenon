package enums

import "strings"

type SecretState string

func (v SecretState) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v SecretState) String() string {
	return string(v)
}
