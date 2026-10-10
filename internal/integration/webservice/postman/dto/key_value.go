package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var keyValueKeys = []string{"key", "value", "disabled", "description"}

type KeyValue struct {
	Raw         map[string]any `json:"-"`
	Key         any            `json:"key"`
	Value       any            `json:"value"`
	Description Description    `json:"description"`
	Disabled    bool           `json:"disabled"`
}

func NewKeyValue(key, value string) KeyValue {
	return KeyValue{Key: key, Value: value}
}

func (k *KeyValue) UnmarshalJSON(data []byte) error {
	type plain KeyValue
	if err := json.Unmarshal(data, (*plain)(k)); err != nil {
		return err
	}
	k.Raw = decodeRaw(data)
	return nil
}

func (k *KeyValue) ToEntity() entity.KeyValue {
	row := entity.NewKeyValue(text(k.Key), text(k.Value))
	row.Enabled = !k.Disabled
	row.Description = k.Description.Content
	return row
}

func (k *KeyValue) ToPathVariable() entity.PathVariable {
	return entity.PathVariable{Key: text(k.Key), Value: text(k.Value), Description: k.Description.Content}
}

func (k *KeyValue) ToVariableRow() entity.VariableRow {
	row := entity.NewVariableRow(text(k.Key), text(k.Value))
	row.Enabled = !k.Disabled
	row.Description = k.Description.Content
	if passthrough := k.Passthrough(); len(passthrough) > 0 {
		row.Postman = passthrough
	}
	return row
}

func (k *KeyValue) Passthrough() map[string]any {
	fields := without(k.Raw, keyValueKeys...)
	return withEntry(fields, "description", k.Description.Passthrough())
}

func keyValuesToEntity(rows []KeyValue) []entity.KeyValue {
	result := make([]entity.KeyValue, 0, len(rows))
	for index := range rows {
		result = append(result, rows[index].ToEntity())
	}
	return result
}

func keyValuePassthrough(rows []KeyValue) []any {
	result := make([]any, 0, len(rows))
	for index := range rows {
		result = append(result, rows[index].Passthrough())
	}
	return result
}
