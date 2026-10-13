package sendoutcome

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Sent                 enums.SendOutcome = "sent"
	SkippedByScript      enums.SendOutcome = "skipped_by_script"
	StoppedByScriptError enums.SendOutcome = "stopped_by_script_error"
	Cancelled            enums.SendOutcome = "cancelled"
	Error                enums.SendOutcome = "error"
	Unknown              enums.SendOutcome = "unknown"
)

func GetSendOutcome(s string) enums.SendOutcome {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sent":
		return Sent
	case "skipped_by_script":
		return SkippedByScript
	case "stopped_by_script_error":
		return StoppedByScriptError
	case "cancelled":
		return Cancelled
	case "error":
		return Error
	default:
		return Unknown
	}
}
