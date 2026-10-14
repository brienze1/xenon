package dto

func NewCreateFolder(localId, parentId, name string) CreateFolder {
	return CreateFolder{LocalId: localId, ParentId: parentId, Name: name}
}
