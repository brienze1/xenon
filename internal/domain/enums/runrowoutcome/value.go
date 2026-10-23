package runrowoutcome

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Passed    enums.RunRowOutcome = "passed"
	Failed    enums.RunRowOutcome = "failed"
	Skipped   enums.RunRowOutcome = "skipped"
	Error     enums.RunRowOutcome = "error"
	Cancelled enums.RunRowOutcome = "cancelled"
	Unknown   enums.RunRowOutcome = "unknown"
)

func GetRunRowOutcome(s string) enums.RunRowOutcome {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "passed":
		return Passed
	case "failed":
		return Failed
	case "skipped":
		return Skipped
	case "error":
		return Error
	case "cancelled":
		return Cancelled
	default:
		return Unknown
	}
}
