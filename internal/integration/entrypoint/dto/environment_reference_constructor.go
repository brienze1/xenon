package dto

func NewEnvironmentReference(localId, id string) EnvironmentReference {
	return EnvironmentReference{LocalId: localId, Id: id}
}
