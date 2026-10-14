package enums

import "strings"

type MCPTokenState string

func (m MCPTokenState) Equals(raw string) bool {
	return strings.EqualFold(string(m), raw)
}

func (m MCPTokenState) String() string {
	return string(m)
}
