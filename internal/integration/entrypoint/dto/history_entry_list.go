package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type HistoryEntryList []HistoryEntry

func NewHistoryEntryList(entries []entity.HistoryEntry) HistoryEntryList {
	list := make(HistoryEntryList, 0, len(entries))
	for _, entry := range entries {
		list = append(list, NewHistoryEntry(entry))
	}
	return list
}

func (h HistoryEntryList) LogValue() any {
	return map[string]any{"entries": len(h)}
}
