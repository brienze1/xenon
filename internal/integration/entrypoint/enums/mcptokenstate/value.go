package mcptokenstate

import (
	"strings"

	"github.com/brienze1/xenon/internal/integration/entrypoint/enums"
)

const (
	Valid   enums.MCPTokenState = "valid"
	Expired enums.MCPTokenState = "expired"
	Unknown enums.MCPTokenState = "unknown"
)

func GetMCPTokenState(s string) enums.MCPTokenState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "valid":
		return Valid
	case "expired":
		return Expired
	default:
		return Unknown
	}
}
