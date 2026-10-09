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

func (k KeyValue) IsEnabled() bool {
	return k.Enabled
}

func (k KeyValue) IsBlank() bool {
	return k.Key == "" && k.Value == ""
}

func (k KeyValue) LogValue() any {
	return map[string]any{"key": k.Key, "enabled": k.Enabled}
}
