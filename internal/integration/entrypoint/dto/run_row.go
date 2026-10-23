package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunRow struct {
	RequestId string `json:"request_id" validate:"required,not_blank"`
	AliasId   string `json:"alias_id"`
	Selected  bool   `json:"selected"`
}

func NewRunRow(row entity.RunRow) RunRow {
	return RunRow{RequestId: row.RequestID, AliasId: row.AliasID, Selected: row.Selected}
}

func (r *RunRow) ToEntity() entity.RunRow {
	return entity.RunRow{RequestID: r.RequestId, AliasID: r.AliasId, Selected: r.Selected}
}
