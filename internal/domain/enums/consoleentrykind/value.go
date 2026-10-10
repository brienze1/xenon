package consoleentrykind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Request     enums.ConsoleEntryKind = "request"
	Log         enums.ConsoleEntryKind = "log"
	Error       enums.ConsoleEntryKind = "error"
	TestFailure enums.ConsoleEntryKind = "test_failure"
	Info        enums.ConsoleEntryKind = "info"
	Unknown     enums.ConsoleEntryKind = "unknown"
)

func GetConsoleEntryKind(s string) enums.ConsoleEntryKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "request":
		return Request
	case "log":
		return Log
	case "error":
		return Error
	case "test_failure":
		return TestFailure
	case "info":
		return Info
	default:
		return Unknown
	}
}
