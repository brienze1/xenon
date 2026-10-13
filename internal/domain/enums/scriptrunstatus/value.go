package scriptrunstatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Ok      enums.ScriptRunStatus = "ok"
	Error   enums.ScriptRunStatus = "error"
	Skipped enums.ScriptRunStatus = "skipped"
	Unknown enums.ScriptRunStatus = "unknown"
)

func GetScriptRunStatus(s string) enums.ScriptRunStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ok":
		return Ok
	case "error":
		return Error
	case "skipped":
		return Skipped
	default:
		return Unknown
	}
}
