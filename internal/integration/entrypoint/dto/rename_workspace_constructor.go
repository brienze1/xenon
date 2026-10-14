package dto

func NewRenameWorkspace(localId, name string) RenameWorkspace {
	return RenameWorkspace{LocalId: localId, Name: name}
}
