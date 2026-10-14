package dto

func NewVariableKeys(localId, scope string, keys []string) VariableKeys {
	return VariableKeys{LocalId: localId, Scope: scope, Keys: keys}
}
