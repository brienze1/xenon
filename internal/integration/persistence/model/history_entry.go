package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

const historyVersion = 1

type HistoryEntry struct {
	SentAt          time.Time     `json:"sent_at"`
	Id              string        `json:"id"`
	SourceRequestId string        `json:"source_request_id"`
	Outcome         string        `json:"outcome"`
	Result          HistoryResult `json:"result"`
	Request         Request       `json:"request"`
	Version         int           `json:"version"`
}

func NewHistoryEntry(entry entity.HistoryEntry) HistoryEntry {
	return HistoryEntry{
		Version:         historyVersion,
		Id:              entry.ID,
		SentAt:          entry.SentAt.UTC(),
		SourceRequestId: entry.SourceRequestID,
		Outcome:         entry.Outcome.String(),
		Request:         NewRequest(entry.Request),
		Result:          NewHistoryResult(entry.Result),
	}
}

func (h *HistoryEntry) ToEntity() entity.HistoryEntry {
	return entity.HistoryEntry{
		ID:              h.Id,
		SentAt:          h.SentAt.UTC(),
		SourceRequestID: h.SourceRequestId,
		Outcome:         historyOutcome(h.Outcome),
		Request:         h.Request.ToEntity(),
		Result:          h.Result.ToEntity(),
	}
}

func historyOutcome(raw string) enums.SendOutcome {
	if raw == "" {
		return sendoutcome.Sent
	}
	return sendoutcome.GetSendOutcome(raw)
}
