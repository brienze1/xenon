package dto

func NewSetActiveEnvironment(localId, id string) SetActiveEnvironment {
	return SetActiveEnvironment{LocalId: localId, Id: id}
}
