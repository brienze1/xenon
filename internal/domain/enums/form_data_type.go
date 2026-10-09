package enums

import "strings"

type FormDataType string

func (f FormDataType) Equals(s string) bool {
	return strings.EqualFold(string(f), s)
}

func (f FormDataType) String() string {
	return string(f)
}
