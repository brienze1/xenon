package dto

func NewRenameItem(localId, id, name string) RenameItem {
	return RenameItem{LocalId: localId, Id: id, Name: name}
}
