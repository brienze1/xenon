package entity

// NewAIHistoryPage keeps 100 of up to 101 fetched entries; Next is the last kept id when more exist.
func NewAIHistoryPage(entries []HistoryEntry) AIPage[HistoryEntry] {
	page := AIPage[HistoryEntry]{Items: make([]HistoryEntry, 0, min(len(entries), AIPageSize))}
	for index, entry := range entries {
		if index == AIPageSize {
			page.Next = entries[index-1].ID
			break
		}
		page.Items = append(page.Items, entry)
	}
	return page
}
