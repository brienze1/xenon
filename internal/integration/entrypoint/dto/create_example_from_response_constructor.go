package dto

func NewCreateExampleFromResponse(localId, requestId, tabId, name string) CreateExampleFromResponse {
	return CreateExampleFromResponse{LocalId: localId, RequestId: requestId, TabId: tabId, Name: name}
}
