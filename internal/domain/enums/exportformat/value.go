package exportformat

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	XenonYaml enums.ExportFormat = "xenon_yaml"
	Postman   enums.ExportFormat = "postman"
	Unknown   enums.ExportFormat = "unknown"
)

func GetExportFormat(s string) enums.ExportFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "xenon_yaml":
		return XenonYaml
	case "postman":
		return Postman
	default:
		return Unknown
	}
}
