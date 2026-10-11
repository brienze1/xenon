package enums

import "strings"

type SyncState string

func (v SyncState) Equals(s string) bool {
	return strings.EqualFold(string(v), s)
}

func (v SyncState) String() string {
	return string(v)
}
