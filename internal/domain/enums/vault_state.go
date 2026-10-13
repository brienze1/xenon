package enums

import "strings"

type VaultState string

func (v VaultState) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v VaultState) String() string {
	return string(v)
}
