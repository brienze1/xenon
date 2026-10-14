package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistorySend struct {
	Status  *int              `json:"status,omitempty"`
	Error   string            `json:"error,omitempty"`
	Request MCPHistoryRequest `json:"request"`
	Ms      int64             `json:"ms"`
	Size    int64             `json:"size"`
}

func NewMCPHistorySend(historyEntry entity.HistoryEntry) MCPHistorySend {
	entry := NewHistoryEntry(historyEntry)
	return MCPHistorySend{
		Request: NewMCPHistoryRequest(historyEntry.Request),
		Status:  entry.StatusCode,
		Ms:      entry.TotalMs,
		Size:    entry.SizeBytes,
		Error:   entry.ErrorKind,
	}
}
