package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPRunRow struct {
	RequestId string `json:"request_id" description:"Request id." validate:"required,not_blank"`
	AliasId   string `json:"alias_id" description:"Alias id; empty for the Default."`
}

func (m *MCPRunRow) ToEntity() entity.RunRow {
	return entity.RunRow{RequestID: m.RequestId, AliasID: m.AliasId, Selected: true}
}
