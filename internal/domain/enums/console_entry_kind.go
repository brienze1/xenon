package enums

import "strings"

type ConsoleEntryKind string

func (v ConsoleEntryKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ConsoleEntryKind) String() string {
	return string(v)
}
