package snippetlanguage

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Curl           enums.SnippetLanguage = "curl"
	JsFetch        enums.SnippetLanguage = "js_fetch"
	GoNative       enums.SnippetLanguage = "go_native"
	PythonRequests enums.SnippetLanguage = "python_requests"
	Unknown        enums.SnippetLanguage = "unknown"
)

func GetSnippetLanguage(s string) enums.SnippetLanguage {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "curl":
		return Curl
	case "js_fetch":
		return JsFetch
	case "go_native":
		return GoNative
	case "python_requests":
		return PythonRequests
	default:
		return Unknown
	}
}
