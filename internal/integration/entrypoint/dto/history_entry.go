package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type HistoryEntry struct {
	SentAt          string `json:"sent_at"`
	StatusCode      *int   `json:"status_code,omitempty"`
	Id              string `json:"id"`
	Method          string `json:"method"`
	Url             string `json:"url"`
	ErrorKind       string `json:"error_kind,omitempty"`
	SourceRequestId string `json:"source_request_id"`
	Outcome         string `json:"outcome"`
	TotalMs         int64  `json:"total_ms"`
	SizeBytes       int64  `json:"size_bytes"`
}

func NewHistoryEntry(entry entity.HistoryEntry) HistoryEntry {
	result := HistoryEntry{
		Id:              entry.ID,
		SentAt:          entry.SentAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Method:          entry.Request.NormalizedMethod(),
		Url:             entry.Request.URL,
		ErrorKind:       entry.Result.ErrorKind.String(),
		SourceRequestId: entry.SourceRequestID,
		Outcome:         entry.Outcome.String(),
		TotalMs:         entry.Result.Total.Milliseconds(),
		SizeBytes:       entry.Result.SizeBytes,
	}
	if entry.HasStatus() {
		statusCode := entry.Result.StatusCode
		result.StatusCode = &statusCode
	}
	return result
}
