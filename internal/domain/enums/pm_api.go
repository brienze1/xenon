package enums

import "strings"

type PmApi string

func (p PmApi) Equals(s string) bool {
	return strings.EqualFold(string(p), s)
}

func (p PmApi) String() string {
	return string(p)
}
