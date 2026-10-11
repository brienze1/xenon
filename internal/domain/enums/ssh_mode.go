package enums

import "strings"

type SshMode string

func (v SshMode) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v SshMode) String() string {
	return string(v)
}
