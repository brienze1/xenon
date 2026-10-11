package enums

import "strings"

type HostKeyStatus string

func (h HostKeyStatus) Equals(s string) bool {
	return strings.EqualFold(string(h), s)
}

func (h HostKeyStatus) String() string {
	return string(h)
}
