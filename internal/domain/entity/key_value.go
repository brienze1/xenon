package entity

import "strings"

type KeyValue struct {
	Key         string
	Value       string
	Description string
	Enabled     bool
}

func NewKeyValue(key, value string) KeyValue {
	return KeyValue{Key: key, Value: value, Enabled: true}
}

func (k KeyValue) IsActive() bool {
	return k.Enabled && strings.TrimSpace(k.Key) != ""
}

func (k KeyValue) HasKey(name string) bool {
	return strings.EqualFold(strings.TrimSpace(k.Key), strings.TrimSpace(name))
}
