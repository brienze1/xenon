package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptKeyValue struct {
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Disabled bool   `json:"disabled,omitempty"`
}

func NewScriptKeyValue(row entity.KeyValue) ScriptKeyValue {
	return ScriptKeyValue{Key: row.Key, Value: row.Value, Disabled: !row.Enabled}
}

func NewScriptKeyValues(rows []entity.KeyValue) []ScriptKeyValue {
	values := make([]ScriptKeyValue, 0, len(rows))
	for _, row := range rows {
		values = append(values, NewScriptKeyValue(row))
	}
	return values
}

func (s ScriptKeyValue) ToEntity() entity.KeyValue {
	return entity.KeyValue{Key: s.Key, Value: scriptText(s.Value), Enabled: !s.Disabled}
}

func (s ScriptKeyValue) Equals(other ScriptKeyValue) bool {
	return s.Key == other.Key && scriptText(s.Value) == scriptText(other.Value) && s.Disabled == other.Disabled
}

func scriptKeyValuesToEntity(values []ScriptKeyValue) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0, len(values))
	for _, value := range values {
		rows = append(rows, value.ToEntity())
	}
	return rows
}

func sameKeyValues(a, b []ScriptKeyValue) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if !a[index].Equals(b[index]) {
			return false
		}
	}
	return true
}

func scriptParameter(values []ScriptKeyValue, key string) string {
	for _, value := range values {
		if value.Key == key {
			return scriptText(value.Value)
		}
	}
	return ""
}

func scriptText(value any) string {
	if value == nil {
		return ""
	}
	return entity.LocalVariableText(value)
}
