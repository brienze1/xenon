package rawlanguage

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Json       enums.RawLanguage = "json"
	Xml        enums.RawLanguage = "xml"
	Text       enums.RawLanguage = "text"
	Html       enums.RawLanguage = "html"
	JavaScript enums.RawLanguage = "javascript"
	Unknown    enums.RawLanguage = "unknown"
)

func GetRawLanguage(s string) enums.RawLanguage {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return Json
	case "xml":
		return Xml
	case "text":
		return Text
	case "html":
		return Html
	case "javascript":
		return JavaScript
	default:
		return Unknown
	}
}
