package enums

import "strings"

type HistorySource string

func (v HistorySource) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v HistorySource) String() string {
	return string(v)
}
