package enums

import "strings"

type DigestAlgorithm string

func (d DigestAlgorithm) Equals(s string) bool {
	return strings.EqualFold(string(d), s)
}

func (d DigestAlgorithm) String() string {
	return string(d)
}
