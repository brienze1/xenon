package historyfilter

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	All  enums.HistoryFilter = "all"
	User enums.HistoryFilter = "user"
	AI   enums.HistoryFilter = "ai"
)

func GetHistoryFilter(s string) enums.HistoryFilter {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "user":
		return User
	case "ai":
		return AI
	default:
		return All
	}
}
