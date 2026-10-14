package dto

func NewCreateEnvironment(localId, name string) CreateEnvironment {
	return CreateEnvironment{LocalId: localId, Name: name}
}
