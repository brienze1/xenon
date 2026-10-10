package enums

import "strings"

type EnvColor string

func (e EnvColor) Equals(s string) bool {
	return strings.EqualFold(string(e), s)
}

func (e EnvColor) String() string {
	return string(e)
}
