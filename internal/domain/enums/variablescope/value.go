package variablescope

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Local       enums.VariableScope = "local"
	Data        enums.VariableScope = "data"
	Environment enums.VariableScope = "environment"
	Collection  enums.VariableScope = "collection"
	Globals     enums.VariableScope = "globals"
	Dynamic     enums.VariableScope = "dynamic"
	None        enums.VariableScope = "none"
	Unknown     enums.VariableScope = "unknown"
)

func GetVariableScope(s string) enums.VariableScope {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "local":
		return Local
	case "data":
		return Data
	case "environment":
		return Environment
	case "collection":
		return Collection
	case "globals":
		return Globals
	case "dynamic":
		return Dynamic
	case "none":
		return None
	default:
		return Unknown
	}
}
