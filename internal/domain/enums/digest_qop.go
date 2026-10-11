package enums

import "strings"

type DigestQop string

func (d DigestQop) Equals(s string) bool {
	return strings.EqualFold(string(d), s)
}

func (d DigestQop) String() string {
	return string(d)
}
