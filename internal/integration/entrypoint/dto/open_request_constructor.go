package dto

func NewOpenRequest(localId, id string) OpenRequest {
	return OpenRequest{LocalId: localId, Id: id}
}
