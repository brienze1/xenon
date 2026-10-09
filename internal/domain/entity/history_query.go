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
