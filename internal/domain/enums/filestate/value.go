package filestate

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Ok      enums.FileState = "ok"
	Repick  enums.FileState = "repick"
	Unknown enums.FileState = "unknown"
)

func GetFileState(s string) enums.FileState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ok":
		return Ok
	case "repick":
		return Repick
	default:
		return Unknown
	}
}
