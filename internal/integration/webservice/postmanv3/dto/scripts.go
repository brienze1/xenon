package dto

import "strings"

type Scripts []Script

func (s Scripts) toPostmanV21(conversion *Conversion, owner, path string, extensions map[string]any) []any {
	events := make([]any, 0, len(s))
	kept := make([]any, 0)
	for _, script := range s {
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
