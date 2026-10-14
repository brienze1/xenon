package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type HistoryEntryDetail struct {
	Activity *AIActivity  `json:"activity,omitempty"`
	Request  Request      `json:"request"`
	Entry    HistoryEntry `json:"entry"`
}

func NewHistoryEntryDetail(entry entity.HistoryEntry) HistoryEntryDetail {
	detail := HistoryEntryDetail{Entry: NewHistoryEntry(entry), Request: NewRequest(entry.Request)}
	if entry.IsAction() {
		activity := NewAIActivity(*entry.Activity, true)
		detail.Activity = &activity
	}
	return detail
}

func (h HistoryEntryDetail) LogValue() any {
	return map[string]any{"id": h.Entry.Id, "request": h.Request.LogValue()}
}
