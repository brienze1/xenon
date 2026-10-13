package dto

const HistoryChangedEvent = "history:changed"

type HistoryChanged struct {
	LocalId string `json:"local_id"`
}

func NewHistoryChanged(localId string) HistoryChanged {
	return HistoryChanged{LocalId: localId}
}
