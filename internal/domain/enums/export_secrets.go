package enums

import "strings"

type ExportSecrets string

func (e ExportSecrets) Equals(s string) bool {
	return strings.EqualFold(string(e), s)
}

func (e ExportSecrets) String() string {
	return string(e)
}
