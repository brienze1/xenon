package scriptmemberkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Object   enums.ScriptMemberKind = "object"
	Function enums.ScriptMemberKind = "function"
	Property enums.ScriptMemberKind = "property"
	Unknown  enums.ScriptMemberKind = "unknown"
)

func GetScriptMemberKind(s string) enums.ScriptMemberKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "object":
		return Object
	case "function":
		return Function
	case "property":
		return Property
	default:
		return Unknown
	}
}
