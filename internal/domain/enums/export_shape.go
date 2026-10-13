package enums

import "strings"

type ExportShape string

func (e ExportShape) Equals(s string) bool {
	return strings.EqualFold(string(e), s)
}

func (e ExportShape) String() string {
	return string(e)
}
