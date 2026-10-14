package dto

func NewCreateCollection(localId, name string) CreateCollection {
	return CreateCollection{LocalId: localId, Name: name}
}
