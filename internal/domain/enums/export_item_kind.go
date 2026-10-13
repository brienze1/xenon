package enums

import "strings"

type ExportItemKind string

func (e ExportItemKind) Equals(s string) bool {
	return strings.EqualFold(string(e), s)
}

func (e ExportItemKind) String() string {
	return string(e)
}
