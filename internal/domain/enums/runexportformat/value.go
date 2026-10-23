package runexportformat

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Json    enums.RunExportFormat = "json"
	Junit   enums.RunExportFormat = "junit"
	Unknown enums.RunExportFormat = "unknown"
)

func GetRunExportFormat(s string) enums.RunExportFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return Json
	case "junit":
		return Junit
	default:
		return Unknown
	}
}
