package enums

import "strings"

type FileState string

func (f FileState) Equals(s string) bool {
	return strings.EqualFold(string(f), s)
}

func (f FileState) String() string {
	return string(f)
}
