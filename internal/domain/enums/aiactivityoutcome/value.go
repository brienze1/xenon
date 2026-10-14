package aiactivityoutcome

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Ok       enums.AIActivityOutcome = "ok"
	Error    enums.AIActivityOutcome = "error"
	Denied   enums.AIActivityOutcome = "denied"
	TimedOut enums.AIActivityOutcome = "timed_out"
	Unknown  enums.AIActivityOutcome = "unknown"
)

func GetAIActivityOutcome(s string) enums.AIActivityOutcome {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ok":
		return Ok
	case "error":
		return Error
	case "denied":
		return Denied
	case "timed_out":
		return TimedOut
	default:
		return Unknown
	}
}
