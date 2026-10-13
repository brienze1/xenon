package snippetphase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Any          enums.SnippetPhase = "any"
	PreRequest   enums.SnippetPhase = "preRequest"
	PostResponse enums.SnippetPhase = "postResponse"
	Unknown      enums.SnippetPhase = "unknown"
)

func GetSnippetPhase(s string) enums.SnippetPhase {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "any":
		return Any
	case "prerequest":
		return PreRequest
	case "postresponse":
		return PostResponse
	default:
		return Unknown
	}
}
