package dto

func NewInspectVariables(localId string, request Request) InspectVariables {
	return InspectVariables{LocalId: localId, Request: request}
}
