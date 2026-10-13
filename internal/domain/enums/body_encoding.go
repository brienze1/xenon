package enums

import "strings"

type BodyEncoding string

func (b BodyEncoding) Equals(s string) bool {
	return strings.EqualFold(string(b), s)
}

func (b BodyEncoding) String() string {
	return string(b)
}
