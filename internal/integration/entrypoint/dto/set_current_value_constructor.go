package dto

func NewSetCurrentValue(localId, scope, key, value string) SetCurrentValue {
	return SetCurrentValue{LocalId: localId, Scope: scope, Key: key, Value: value}
}
