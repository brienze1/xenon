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
	Outcome         string         `json:"outcome,omitempty"`
	Version         int            `json:"version"`
}

type HistorySource struct {
	Source string `json:"source"`
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
	model.Outcome = entry.Outcome.String()
	model.Request = &request
	model.Result = &result
	return model
}

func (h *HistoryEntry) ToEntity() entity.HistoryEntry {
	entry := entity.HistoryEntry{
		ID:     h.Id,
		SentAt: h.SentAt.UTC(),
		Kind:   historykind.GetHistoryKind(h.Kind),
		Source: historysource.GetHistorySource(h.Source),
		Client: h.Client,
	}
	if entry.Kind == historykind.Action && h.Activity != nil {
		activity := h.Activity.ToEntity()
		entry.Activity = &activity
		return entry
	}
	entry.Kind = historykind.Send
	entry.SourceRequestID = h.SourceRequestId
	entry.Outcome = historyOutcome(h.Outcome)
	if h.Request != nil {
		entry.Request = h.Request.ToEntity()
	}
	if h.Result != nil {
		entry.Result = h.Result.ToEntity()
	}
	return entry
}

func (h HistorySource) ToEntity() enums.HistorySource {
	return historysource.GetHistorySource(h.Source)
}

func historyOutcome(raw string) enums.SendOutcome {
	if raw == "" {
		return sendoutcome.Sent
	}
	return sendoutcome.GetSendOutcome(raw)
}
