package dto

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
)

type ClearHistory struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Source  string `json:"source" validate:"omitempty,oneof=all user ai"`
}

func NewClearHistory(localId, source string) ClearHistory {
	return ClearHistory{LocalId: localId, Source: source}
}

func (c *ClearHistory) ToEntity() enums.HistoryFilter {
	return historyfilter.GetHistoryFilter(c.Source)
}
