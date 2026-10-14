package dto

func NewSaveRequest(localId, baseRevision string, request Request) SaveRequest {
	return SaveRequest{LocalId: localId, BaseRevision: baseRevision, Request: request}
}
