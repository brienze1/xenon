package runstatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Running        enums.RunStatus = "running"
	Paused         enums.RunStatus = "paused"
	Completed      enums.RunStatus = "completed"
	Stopped        enums.RunStatus = "stopped"
	StoppedOnError enums.RunStatus = "stopped_on_error"
	Unknown        enums.RunStatus = "unknown"
)

func GetRunStatus(s string) enums.RunStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "running":
		return Running
	case "paused":
		return Paused
	case "completed":
		return Completed
	case "stopped":
		return Stopped
	case "stopped_on_error":
		return StoppedOnError
	default:
		return Unknown
	}
}
