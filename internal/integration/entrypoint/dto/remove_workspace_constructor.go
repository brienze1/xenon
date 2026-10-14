package dto

func NewRemoveWorkspace(localId string, moveToTrash bool) RemoveWorkspace {
	return RemoveWorkspace{LocalId: localId, MoveToTrash: moveToTrash}
}
