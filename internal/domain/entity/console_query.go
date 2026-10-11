package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
)

type ConsoleQuery struct {
	Filter  enums.ConsoleFilter
	AfterID string
	Search  string
	Limit   int
}

func (q ConsoleQuery) Matches(entry ConsoleEntry) bool {
	return q.matchesFilter(entry) && entry.Contains(q.Search)
}

func (q ConsoleQuery) matchesFilter(entry ConsoleEntry) bool {
	switch q.Filter {
	case consolefilter.Requests:
		return entry.IsRequest()
	case consolefilter.Logs:
		return entry.IsLog()
	case consolefilter.Errors:
		return entry.IsError()
	default:
		return true
	}
}

func searchable(text, search string) bool {
	return strings.Contains(strings.ToLower(text), search)
}
