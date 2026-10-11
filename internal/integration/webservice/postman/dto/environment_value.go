package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	secretVariableType  = "secret"
	defaultVariableType = "default"
	stringVariableType  = "string"
	numberVariableType  = "number"
	booleanVariableType = "boolean"
)

var environmentValueKeys = []string{"key", "value", "enabled"}

type EnvironmentValue struct {
	Raw     map[string]any `json:"-"`
	Enabled *bool          `json:"enabled"`
	Key     any            `json:"key"`
	Value   any            `json:"value"`
	Type    string         `json:"type"`
}

func (e *EnvironmentValue) UnmarshalJSON(data []byte) error {
	type plain EnvironmentValue
	if err := json.Unmarshal(data, (*plain)(e)); err != nil {
		return err
	}
	e.Raw = decodeRaw(data)
	return nil
}

func (e *EnvironmentValue) IsSecret() bool {
	return e.Type == secretVariableType
}

func (e *EnvironmentValue) ToEntity() entity.VariableRow {
	row := entity.NewVariableRow(text(e.Key), text(e.Value))
	row.Enabled = e.Enabled == nil || *e.Enabled
	row.Secret = e.IsSecret()
	if passthrough := without(e.Raw, environmentValueKeys...); len(passthrough) > 0 {
		row.Postman = passthrough
	}
	return row.WithoutSecretValue()
}

func NewEnvironmentValue(row entity.VariableRow) EnvironmentValue {
	passthrough := map[string]any(row.Postman)
	enabled := row.Enabled
	value := EnvironmentValue{Key: row.Key, Value: row.InitialValue, Type: text(passthrough["type"]), Enabled: &enabled, Raw: without(passthrough, "type")}
	switch {
	case row.Secret:
		value.Type, value.Value = secretVariableType, ""
	case value.Type == "" || value.Type == secretVariableType:
		value.Type = defaultVariableType
	}
	if row.Description != "" {
		value.Raw["description"] = row.Description
	}
	return value
}

func (e EnvironmentValue) MarshalJSON() ([]byte, error) {
	value := newObject().set("key", e.Key).set("value", e.Value).set("type", e.Type).set("enabled", e.Enabled != nil && *e.Enabled)
	return value.merged(e.Raw).MarshalJSON()
}
