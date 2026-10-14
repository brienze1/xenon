package historysource

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	User    enums.HistorySource = "user"
	AI      enums.HistorySource = "ai"
	Unknown enums.HistorySource = "unknown"
)

func GetHistorySource(s string) enums.HistorySource {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "user", "":
		return User
	case "ai":
		return AI
	default:
		return Unknown
	}
}
