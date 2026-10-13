package consolelevel

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Log     enums.ConsoleLevel = "log"
	Info    enums.ConsoleLevel = "info"
	Warn    enums.ConsoleLevel = "warn"
	Error   enums.ConsoleLevel = "error"
	Debug   enums.ConsoleLevel = "debug"
	Clear   enums.ConsoleLevel = "clear"
	Unknown enums.ConsoleLevel = "unknown"
)

func GetConsoleLevel(s string) enums.ConsoleLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "log":
		return Log
	case "info":
		return Info
	case "warn":
		return Warn
	case "error":
		return Error
	case "debug":
		return Debug
	case "clear":
		return Clear
	default:
		return Unknown
	}
}
