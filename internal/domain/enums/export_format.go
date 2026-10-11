package enums

import "strings"

type ExportFormat string

func (e ExportFormat) Equals(s string) bool {
	return strings.EqualFold(string(e), s)
}

func (e ExportFormat) String() string {
	return string(e)
}
