package scripterrorkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Syntax             enums.ScriptErrorKind = "script_syntax"
	Runtime            enums.ScriptErrorKind = "script_runtime"
	Timeout            enums.ScriptErrorKind = "script_timeout"
	Memory             enums.ScriptErrorKind = "script_memory"
	StackOverflow      enums.ScriptErrorKind = "script_stack_overflow"
	UnhandledRejection enums.ScriptErrorKind = "script_unhandled_rejection"
	UnsupportedApi     enums.ScriptErrorKind = "script_unsupported_api"
	Unknown            enums.ScriptErrorKind = "unknown"
)

func GetScriptErrorKind(s string) enums.ScriptErrorKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "script_syntax":
		return Syntax
	case "script_runtime":
		return Runtime
	case "script_timeout":
		return Timeout
	case "script_memory":
		return Memory
	case "script_stack_overflow":
		return StackOverflow
	case "script_unhandled_rejection":
		return UnhandledRejection
	case "script_unsupported_api":
		return UnsupportedApi
	default:
		return Unknown
	}
}
