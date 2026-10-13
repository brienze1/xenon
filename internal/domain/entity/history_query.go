package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
)

type HistoryQuery struct {
	Source   enums.HistoryFilter
	Text     string
	BeforeID string
	Limit    int
}

func NewHistoryQuery(text, beforeId string, limit int) HistoryQuery {
	if limit == 0 {
		limit = DefaultHistoryPageSize
	}
	return HistoryQuery{Source: historyfilter.All, Text: text, BeforeID: beforeId, Limit: limit}
}

func (h HistoryQuery) FromSource(source enums.HistoryFilter) HistoryQuery {
	h.Source = source
	return h
}

func (h HistoryQuery) IsFilledBy(count int) bool {
	return count >= h.Limit
}

func (h HistoryQuery) Matches(entry HistoryEntry) bool {
	return entry.IsIn(h.Source) && entry.Matches(h.Text)
}

func (h HistoryQuery) HasCursor() bool {
	return h.BeforeID != ""
}

func (h HistoryQuery) IsCursor(entry HistoryEntry) bool {
	return entry.HasID(h.BeforeID)
}
