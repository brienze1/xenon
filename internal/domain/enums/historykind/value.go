package historykind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Send    enums.HistoryKind = "send"
	Action  enums.HistoryKind = "action"
	Unknown enums.HistoryKind = "unknown"
)

func GetHistoryKind(s string) enums.HistoryKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "send", "":
		return Send
	case "action":
		return Action
	default:
		return Unknown
	}
}
