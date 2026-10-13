package aiaccessstate

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Off     enums.AIAccessState = "off"
	Running enums.AIAccessState = "running"
	Error   enums.AIAccessState = "error"
)

func GetAIAccessState(s string) enums.AIAccessState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "running":
		return Running
	case "error":
		return Error
	default:
		return Off
	}
}
