package dto

import (
	"encoding/json"
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var keyValueKeys = []string{"key", "value", "disabled", "description"}

type KeyValue struct {
	Raw         map[string]any `json:"-"`
	Key         any            `json:"key"`
	Value       any            `json:"value"`
	Type        string         `json:"type"`
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

func (k *KeyValue) IsSecret() bool {
	return k.Type == secretVariableType
}

func (k *KeyValue) ToVariableRow() entity.VariableRow {
	row := entity.NewVariableRow(text(k.Key), text(k.Value))
	row.Secret = k.IsSecret()
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

func NewKeyValueRow(row entity.KeyValue, passthrough map[string]any) KeyValue {
	return KeyValue{Key: row.Key, Value: row.Value, Disabled: !row.Enabled, Description: newDescription(row.Description, passthrough["description"]), Raw: passthrough}
}

func NewPathVariableRow(variable entity.PathVariable, passthrough map[string]any) KeyValue {
	return KeyValue{Key: variable.Key, Value: variable.Value, Description: newDescription(variable.Description, passthrough["description"]), Raw: passthrough}
}

func NewVariable(row entity.VariableRow) KeyValue {
	passthrough := map[string]any(row.Postman)
	variable := KeyValue{Key: row.Key, Value: row.InitialValue, Type: text(passthrough["type"]), Disabled: !row.Enabled, Description: newDescription(row.Description, passthrough["description"]), Raw: passthrough}
	switch {
	case row.Secret:
		variable.Type, variable.Value = stringVariableType, ""
	case variable.Type == secretVariableType:
		variable.Type = stringVariableType
	default:
		variable.Value = typedValue(variable.Type, row.InitialValue)
	}
	return variable
}

func (k KeyValue) MarshalJSON() ([]byte, error) {
	row := newObject().set("key", k.Key).set("value", k.Value).
		setIf(k.Type != "", "type", k.Type).
		setIf(k.Disabled, "disabled", true).
		setPresent("description", k.Description.Value())
	return row.merged(k.Raw).MarshalJSON()
}

func typedValue(variableType, value string) any {
	switch variableType {
	case numberVariableType:
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			return number
		}
	case booleanVariableType:
		if flag, err := strconv.ParseBool(value); err == nil {
			return flag
		}
	}
	return value
}
