package runtype

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Functional  enums.RunType = "functional"
	Performance enums.RunType = "performance"
	Unknown     enums.RunType = "unknown"
	Unset       enums.RunType = ""
)

func GetRunType(s string) enums.RunType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "functional":
		return Functional
	case "performance":
		return Performance
	default:
		return Unknown
	}
}
