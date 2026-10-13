package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ListHistory struct {
	LocalId  string `json:"local_id" validate:"required,not_blank"`
	Query    string `json:"query"`
	BeforeId string `json:"before_id,omitempty"`
	Limit    int    `json:"limit" validate:"min=0,max=1000"`
}

func (l *ListHistory) ToEntity() entity.HistoryQuery {
	return entity.NewHistoryQuery(l.Query, l.BeforeId, l.Limit)
}
