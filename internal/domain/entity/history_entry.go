package entity

import (
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

type HistoryEntry struct {
	SentAt          time.Time
	ID              string
	SourceRequestID string
	Outcome         enums.SendOutcome
	Result          HistoryResult
	Request         Request
}

// NewHistoryEntry records the request as written; a send stopped by a script has no result.
func NewHistoryEntry(id string, sentAt time.Time, request Request, execution SendExecution) HistoryEntry {
	entry := HistoryEntry{
		ID:              id,
		SentAt:          sentAt,
		Request:         request,
		SourceRequestID: request.ID,
		Outcome:         execution.Outcome,
	}
	if execution.Response != nil {
		entry.Result = NewHistoryResult(*execution.Response)
	}
	return entry
}

func (h HistoryEntry) HasStatus() bool {
	return h.Result.IsOk() && h.WentOut()
}

func (h HistoryEntry) WentOut() bool {
	return h.Outcome != sendoutcome.SkippedByScript && h.Outcome != sendoutcome.StoppedByScriptError
}

func (h HistoryEntry) Matches(query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	return strings.Contains(strings.ToLower(h.Request.NormalizedMethod()), needle) ||
		strings.Contains(strings.ToLower(h.Request.URL), needle)
}

func (h HistoryEntry) HasID(id string) bool {
	return h.ID == id
}

func (h HistoryEntry) LogValue() any {
	return map[string]any{"id": h.ID, "request": h.Request.LogValue(), "statusCode": h.Result.StatusCode}
}
