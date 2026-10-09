package entity

import (
	"strings"
	"time"
)

type HistoryEntry struct {
	SentAt          time.Time
	ID              string
	SourceRequestID string
	Request         Request
	Result          HistoryResult
}

func NewHistoryEntry(id string, sentAt time.Time, request Request, response Response) HistoryEntry {
	return HistoryEntry{
		ID:              id,
		SentAt:          sentAt,
		Request:         request,
		SourceRequestID: request.ID,
		Result:          NewHistoryResult(response),
	}
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
