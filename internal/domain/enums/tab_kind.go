package enums

import "strings"

type TabKind string

func (t TabKind) Equals(raw string) bool {
	return strings.EqualFold(string(t), raw)
}

func (t TabKind) String() string {
	return string(t)
}
