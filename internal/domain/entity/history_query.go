package entity

type HistoryQuery struct {
	Text     string
	BeforeID string
	Limit    int
}

func NewHistoryQuery(text, beforeId string, limit int) HistoryQuery {
	if limit == 0 {
		limit = HistoryCap
	}
	return HistoryQuery{Text: text, BeforeID: beforeId, Limit: limit}
}

func (h HistoryQuery) IsFilledBy(count int) bool {
	return count >= h.Limit
}

func (h HistoryQuery) Matches(entry HistoryEntry) bool {
	return entry.Matches(h.Text)
}

func (h HistoryQuery) HasCursor() bool {
	return h.BeforeID != ""
}

func (h HistoryQuery) IsCursor(entry HistoryEntry) bool {
	return entry.HasID(h.BeforeID)
}
