package entity

import (
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
	"github.com/brienze1/xenon/internal/domain/enums/historykind"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

type HistoryEntry struct {
	SentAt          time.Time
	Activity        *AIActivity
	ID              string
	Kind            enums.HistoryKind
	Source          enums.HistorySource
	Client          string
	SourceRequestID string
	AliasID         string
	AliasName       string
	Outcome         enums.SendOutcome
	Result          HistoryResult
	Request         Request
}

// NewHistoryEntry records the request as written; a send stopped by a script has no result.
func NewHistoryEntry(id string, sentAt time.Time, request Request, execution SendExecution, origin SendOrigin) HistoryEntry {
	normalized := origin.Normalized()
	entry := HistoryEntry{
		ID:              id,
		SentAt:          sentAt,
		Kind:            historykind.Send,
		Source:          normalized.Source,
		Client:          normalized.Client,
		Request:         request,
		SourceRequestID: request.ID,
		Outcome:         execution.Outcome,
	}
	if execution.Response != nil {
		entry.Result = NewHistoryResult(*execution.Response)
	}
	return entry
}

func NewActionHistoryEntry(id string, at time.Time, client string, activity AIActivity) HistoryEntry {
	return HistoryEntry{
		ID:       id,
		SentAt:   at,
		Kind:     historykind.Action,
		Source:   historysource.AI,
		Client:   AIClientName(client),
		Activity: &activity,
	}
}

// ForAlias records the alias a send used.
func (h HistoryEntry) ForAlias(alias RequestAlias) HistoryEntry {
	h.AliasID = alias.ID
	h.AliasName = alias.Name
	return h
}

func (h HistoryEntry) IsAction() bool {
	return h.Kind == historykind.Action && h.Activity != nil
}

// Normalized reads an entry of unknown kind, or an action without its activity, as a send.
func (h HistoryEntry) Normalized() HistoryEntry {
	if h.IsAction() {
		return HistoryEntry{ID: h.ID, SentAt: h.SentAt, Kind: h.Kind, Source: h.Source, Client: h.Client, Activity: h.Activity}
	}
	h.Kind = historykind.Send
	h.Activity = nil
	return h
}

func (h HistoryEntry) Origin() enums.HistorySource {
	return HistoryOrigin(h.Source)
}

func HistoryOrigin(source enums.HistorySource) enums.HistorySource {
	if source == historysource.AI {
		return historysource.AI
	}
	return historysource.User
}

func (h HistoryEntry) IsIn(filter enums.HistoryFilter) bool {
	switch filter {
	case historyfilter.User:
		return h.Origin() == historysource.User
	case historyfilter.AI:
		return h.Origin() == historysource.AI
	default:
		return true
	}
}

func (h HistoryEntry) HasStatus() bool {
	return h.Result.IsOk() && h.WentOut()
}

func (h HistoryEntry) WentOut() bool {
	return h.Outcome != sendoutcome.SkippedByScript && h.Outcome != sendoutcome.StoppedByScriptError
}

func (h HistoryEntry) Matches(query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if h.IsAction() {
		return h.Activity.Mentions(needle)
	}
	return strings.Contains(strings.ToLower(h.Request.NormalizedMethod()), needle) ||
		strings.Contains(strings.ToLower(h.Request.URL), needle)
}

func (h HistoryEntry) HasID(id string) bool {
	return h.ID == id
}

func (h HistoryEntry) ActivityTarget() AIActivityTarget {
	return AIActivityTarget{ID: h.ID}
}

func (h HistoryEntry) LogValue() any {
	if h.IsAction() {
		return map[string]any{"id": h.ID, "kind": h.Kind, "source": h.Source, "client": h.Client, "activity": h.Activity.LogValue()}
	}
	return map[string]any{"id": h.ID, "kind": h.Kind, "source": h.Source, "request": h.Request.LogValue(), "statusCode": h.Result.StatusCode}
}
