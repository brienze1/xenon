package entity

type HistoryEntries []HistoryEntry

func (h HistoryEntries) LogValue() any {
	values := make([]any, 0, len(h))
	for _, entry := range h {
		values = append(values, entry.LogValue())
	}
	return values
}
