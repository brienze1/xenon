package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const historyVersion = 1

type HistoryEntry struct {
	SentAt          time.Time     `json:"sent_at"`
	Id              string        `json:"id"`
	SourceRequestId string        `json:"source_request_id"`
	Request         Request       `json:"request"`
	Result          HistoryResult `json:"result"`
	Version         int           `json:"version"`
}

func NewHistoryEntry(entry entity.HistoryEntry) HistoryEntry {
	return HistoryEntry{
		Version:         historyVersion,
		Id:              entry.ID,
		SentAt:          entry.SentAt.UTC(),
		SourceRequestId: entry.SourceRequestID,
		Request:         NewRequest(entry.Request),
		Result:          NewHistoryResult(entry.Result),
	}
}

func (h *HistoryEntry) ToEntity() entity.HistoryEntry {
	return entity.HistoryEntry{
		ID:              h.Id,
		SentAt:          h.SentAt.UTC(),
		SourceRequestID: h.SourceRequestId,
		Request:         h.Request.ToEntity(),
		Result:          h.Result.ToEntity(),
	}
}
