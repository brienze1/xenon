package dto

func NewAddVariable(localId, scope, key, value, description string, secret bool) AddVariable {
	return AddVariable{LocalId: localId, Scope: scope, Key: key, Value: value, Description: description, Secret: secret}
}
