package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
)

type ListConsole struct {
	AfterId string `json:"after_id"`
	Filter  string `json:"filter" validate:"omitempty,oneof=all requests logs errors"`
	Search  string `json:"search"`
	Limit   int    `json:"limit" validate:"required,min=1,max=5000"`
}

func (l *ListConsole) ToEntity() entity.ConsoleQuery {
	return entity.ConsoleQuery{AfterID: l.AfterId, Limit: l.Limit, Filter: consolefilter.GetConsoleFilter(l.Filter), Search: l.Search}
}
