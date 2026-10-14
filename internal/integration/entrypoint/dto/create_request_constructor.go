package dto

func NewCreateRequest(localId, parentId, name string) CreateRequest {
	return CreateRequest{LocalId: localId, ParentId: parentId, Name: name}
}
