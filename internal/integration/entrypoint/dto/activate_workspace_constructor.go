package dto

func NewActivateWorkspace(localId string) ActivateWorkspace {
	return ActivateWorkspace{LocalId: localId}
}
