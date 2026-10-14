package dto

func NewTreeItemReference(localId, id string) TreeItemReference {
	return TreeItemReference{LocalId: localId, Id: id}
}
