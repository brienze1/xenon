package enums

import "strings"

type HistoryFilter string

func (v HistoryFilter) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v HistoryFilter) String() string {
	return string(v)
}
