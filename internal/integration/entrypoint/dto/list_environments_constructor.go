package dto

func NewListEnvironments(localId string) ListEnvironments {
	return ListEnvironments{LocalId: localId}
}
