package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

type HistorySource struct {
	Source string `json:"source"`
}

func (h HistorySource) ToEntity() enums.HistorySource {
	return entity.HistoryOrigin(historysource.GetHistorySource(h.Source))
}
