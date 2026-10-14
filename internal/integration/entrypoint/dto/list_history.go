package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
)

type ListHistory struct {
	LocalId  string `json:"local_id" validate:"required,not_blank"`
	Source   string `json:"source" validate:"omitempty,oneof=all user ai"`
	Query    string `json:"query"`
	BeforeId string `json:"before_id,omitempty"`
	Limit    int    `json:"limit" validate:"min=0,max=1000"`
}

func NewListHistory(localId, source, query, beforeId string, limit int) ListHistory {
	return ListHistory{LocalId: localId, Source: source, Query: query, BeforeId: beforeId, Limit: limit}
}

func (l *ListHistory) ToEntity() entity.HistoryQuery {
	return entity.NewHistoryQuery(l.Query, l.BeforeId, l.Limit).FromSource(historyfilter.GetHistoryFilter(l.Source))
}
