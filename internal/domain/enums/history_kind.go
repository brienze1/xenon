package enums

import "strings"

type HistoryKind string

func (v HistoryKind) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v HistoryKind) String() string {
	return string(v)
}
