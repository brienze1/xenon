package dto

func NewHistoryReference(localId, id string) HistoryReference {
	return HistoryReference{LocalId: localId, Id: id}
}
