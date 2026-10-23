package enums

import "strings"

type RunExportFormat string

func (v RunExportFormat) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v RunExportFormat) String() string {
	return string(v)
}
