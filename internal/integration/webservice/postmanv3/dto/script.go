package dto

import "strings"

const (
	preRequestListen   = "prerequest"
	postResponseListen = "test"
	scriptMimeType     = "text/javascript"
)

var scriptListens = map[string]string{
	"beforeRequest":      preRequestListen,
	"http:beforeRequest": preRequestListen,
	"beforeQuery":        preRequestListen,
	"afterResponse":      postResponseListen,
	"http:afterResponse": postResponseListen,
}

type Script struct {
	Extra    map[string]any `yaml:",inline"`
	Type     string         `yaml:"type"`
	Code     string         `yaml:"code"`
	Language string         `yaml:"language"`
}

func scriptsToPostmanV21(scripts []Script, conversion *Conversion, owner, path string, extensions map[string]any) []any {
	events := make([]any, 0, len(scripts))
	kept := make([]any, 0)
	for _, script := range scripts {
		listen, known := scriptListens[script.Type]
		if !known {
			conversion.protocolScript(script.Type, owner, path)
			kept = append(kept, map[string]any{"type": script.Type, "code": script.Code, "language": script.Language})
			continue
		}
		events = append(events, map[string]any{"listen": listen, "script": map[string]any{"type": scriptMimeType, "exec": toAnySlice(strings.Split(script.Code, "\n"))}})
	}
	if len(kept) > 0 {
		extensions["scripts"] = kept
	}
	return events
}

func toAnySlice(lines []string) []any {
	result := make([]any, 0, len(lines))
	for _, line := range lines {
		result = append(result, line)
	}
	return result
}
