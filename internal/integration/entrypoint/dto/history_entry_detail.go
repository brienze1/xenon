package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type HistoryEntryDetail struct {
	Request Request      `json:"request"`
	Entry   HistoryEntry `json:"entry"`
}

func NewHistoryEntryDetail(entry entity.HistoryEntry) HistoryEntryDetail {
	return HistoryEntryDetail{Entry: NewHistoryEntry(entry), Request: NewRequest(entry.Request)}
}

func (h HistoryEntryDetail) LogValue() any {
	return map[string]any{"id": h.Entry.Id, "request": h.Request.LogValue()}
}
