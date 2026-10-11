package tokenplacement

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Header      enums.TokenPlacement = "header"
	QueryParams enums.TokenPlacement = "queryParams"
	Unknown     enums.TokenPlacement = "unknown"
)

func GetTokenPlacement(s string) enums.TokenPlacement {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "header":
		return Header
	case "queryparams":
		return QueryParams
	default:
		return Unknown
	}
}
