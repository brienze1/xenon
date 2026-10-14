package dto

func NewMoveItem(localId, id, newParentId string, index int) MoveItem {
	return MoveItem{LocalId: localId, Id: id, NewParentId: newParentId, Index: index}
}
