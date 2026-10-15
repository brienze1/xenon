package dto

func NewReorderEnvironment(localId, id string, index int) ReorderEnvironment {
	return ReorderEnvironment{LocalId: localId, Id: id, Index: index}
}
