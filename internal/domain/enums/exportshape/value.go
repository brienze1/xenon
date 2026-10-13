package exportshape

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Zip     enums.ExportShape = "zip"
	Bundle  enums.ExportShape = "bundle"
	File    enums.ExportShape = "file"
	Unknown enums.ExportShape = "unknown"
)

func GetExportShape(s string) enums.ExportShape {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "zip":
		return Zip
	case "bundle":
		return Bundle
	case "file":
		return File
	default:
		return Unknown
	}
}
