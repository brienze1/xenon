package consolefilter

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	All      enums.ConsoleFilter = "all"
	Requests enums.ConsoleFilter = "requests"
	Logs     enums.ConsoleFilter = "logs"
	Errors   enums.ConsoleFilter = "errors"
)

func GetConsoleFilter(s string) enums.ConsoleFilter {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "requests":
		return Requests
	case "logs":
		return Logs
	case "errors":
		return Errors
	default:
		return All
	}
}
