package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historykind"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

const historyVersion = 1

type HistoryEntry struct {
	SentAt          time.Time      `json:"sent_at"`
	Activity        *AIActivity    `json:"activity,omitempty"`
	Result          *HistoryResult `json:"result,omitempty"`
	Request         *Request       `json:"request,omitempty"`
	Id              string         `json:"id"`
	Kind            string         `json:"kind,omitempty"`
	Source          string         `json:"source,omitempty"`
	Client          string         `json:"client,omitempty"`
	SourceRequestId string         `json:"source_request_id,omitempty"`
	AliasId         string         `json:"alias_id,omitempty"`
	AliasName       string         `json:"alias_name,omitempty"`
	Outcome         string         `json:"outcome,omitempty"`
	Version         int            `json:"version"`
}

func NewHistoryEntry(entry entity.HistoryEntry) HistoryEntry {
	model := HistoryEntry{
		Version: historyVersion,
		Id:      entry.ID,
		SentAt:  entry.SentAt.UTC(),
		Kind:    entry.Kind.String(),
		Source:  entry.Origin().String(),
		Client:  entry.Client,
	}
	if entry.IsAction() {
		activity := NewAIActivity(*entry.Activity)
		model.Activity = &activity
		return model
	}
	request := NewRequest(entry.Request)
	result := NewHistoryResult(entry.Result)
	model.SourceRequestId = entry.SourceRequestID
	model.AliasId = entry.AliasID
	model.AliasName = entry.AliasName
	model.Outcome = entry.Outcome.String()
	model.Request = &request
	model.Result = &result
	return model
}

func (h *HistoryEntry) ToEntity() entity.HistoryEntry {
	entry := entity.HistoryEntry{
		ID:              h.Id,
		SentAt:          h.SentAt.UTC(),
		Kind:            historykind.GetHistoryKind(h.Kind),
		Source:          entity.HistoryOrigin(historysource.GetHistorySource(h.Source)),
		Client:          h.Client,
		SourceRequestID: h.SourceRequestId,
		AliasID:         h.AliasId,
		AliasName:       h.AliasName,
		Outcome:         historyOutcome(h.Outcome),
	}
	if h.Activity != nil {
		activity := h.Activity.ToEntity()
		entry.Activity = &activity
	}
	if h.Request != nil {
		entry.Request = h.Request.ToEntity()
	}
	if h.Result != nil {
		entry.Result = h.Result.ToEntity()
	}
	return entry.Normalized()
}

func historyOutcome(raw string) enums.SendOutcome {
	if raw == "" {
		return sendoutcome.Sent
	}
	return sendoutcome.GetSendOutcome(raw)
}
