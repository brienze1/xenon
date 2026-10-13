package enums

import "strings"

type Divergence string

func (d Divergence) Equals(s string) bool {
	return strings.EqualFold(string(d), s)
}

func (d Divergence) String() string {
	return string(d)
}
