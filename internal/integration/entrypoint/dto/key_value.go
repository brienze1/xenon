package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type KeyValue struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func NewKeyValue(row entity.KeyValue) KeyValue {
	return KeyValue{Key: row.Key, Value: row.Value, Enabled: &row.Enabled, Description: row.Description}
}

func NewKeyValues(rows []entity.KeyValue) []KeyValue {
	result := make([]KeyValue, 0, len(rows))
	for _, row := range rows {
		result = append(result, NewKeyValue(row))
	}
	return result
}

func (k *KeyValue) ToEntity() entity.KeyValue {
	return entity.KeyValue{Key: k.Key, Value: k.Value, Enabled: isEnabled(k.Enabled), Description: k.Description}
}

func keyValuesToEntity(rows []KeyValue) []entity.KeyValue {
	result := make([]entity.KeyValue, 0, len(rows))
	for index := range rows {
		result = append(result, rows[index].ToEntity())
	}
	return result
}

func isEnabled(enabled *bool) bool {
	return enabled == nil || *enabled
}
