package entity

const (
	HistoryCap                 = 1000
	HistoryCompactionThreshold = 1200
)

type History struct {
	entries []HistoryEntry
}

func RetainHistory[T any](oldestFirst []T) ([]T, bool) {
	if len(oldestFirst) <= HistoryCompactionThreshold {
		return oldestFirst, false
	}
	return oldestFirst[len(oldestFirst)-HistoryCap:], true
}

func NewHistory(oldestFirst []HistoryEntry) History {
	kept := oldestFirst[max(len(oldestFirst)-HistoryCap, 0):]
	entries := make([]HistoryEntry, 0, len(kept))
	for index := len(kept) - 1; index >= 0; index-- {
		entries = append(entries, kept[index])
	}
	return History{entries: entries}
}

func (h History) Page(query HistoryQuery) []HistoryEntry {
	page := make([]HistoryEntry, 0, min(query.Limit, len(h.entries)))
	for _, entry := range h.entries[h.startAfter(query.BeforeID):] {
		if len(page) == query.Limit {
			break
		}
		if entry.Matches(query.Text) {
			page = append(page, entry)
		}
	}
	return page
}

func (h History) Find(id string) (HistoryEntry, bool) {
	for _, entry := range h.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return HistoryEntry{}, false
}

func (h History) startAfter(id string) int {
	if id == "" {
		return 0
	}
	for index, entry := range h.entries {
		if entry.ID == id {
			return index + 1
		}
	}
	return len(h.entries)
}
