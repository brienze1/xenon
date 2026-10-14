package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistoryList struct {
	Source string `json:"source"`
	Query  string `json:"query"`
	Cursor string `json:"cursor"`
}

func (m *MCPHistoryList) ListHistory(localId string) ListHistory {
	return NewListHistory(localId, m.Source, m.Query, m.Cursor, entity.AIPageSize+1)
}
