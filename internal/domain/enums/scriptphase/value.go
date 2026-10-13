package scriptphase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	PreRequest   enums.ScriptPhase = "preRequest"
	PostResponse enums.ScriptPhase = "postResponse"
	Unknown      enums.ScriptPhase = "unknown"
)

func GetScriptPhase(s string) enums.ScriptPhase {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "prerequest":
		return PreRequest
	case "postresponse":
		return PostResponse
	default:
		return Unknown
	}
}
