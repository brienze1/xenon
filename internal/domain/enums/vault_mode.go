package enums

import "strings"

type VaultMode string

func (v VaultMode) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v VaultMode) String() string {
	return string(v)
}
