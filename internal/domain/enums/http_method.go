package enums

import "strings"

type HttpMethod string

func (h HttpMethod) Equals(s string) bool {
	return strings.EqualFold(string(h), s)
}

func (h HttpMethod) String() string {
	return string(h)
}
