package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistoryList struct {
	Source string `json:"source" description:"Entries to list; default all." enum:"all,user,ai"`
	Query  string `json:"query" description:"Text the entries must contain."`
	Cursor string `json:"cursor" description:"Cursor from the previous page."`
}

func (m *MCPHistoryList) ListHistory(localId string) ListHistory {
	return NewListHistory(localId, m.Source, m.Query, m.Cursor, entity.AIPageSize+1)
}
