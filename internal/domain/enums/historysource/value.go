package historysource

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	User enums.HistorySource = "user"
	AI   enums.HistorySource = "ai"
)

func GetHistorySource(s string) enums.HistorySource {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ai":
		return AI
	default:
		return User
	}
}
