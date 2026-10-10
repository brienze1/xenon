package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const secretVariableType = "secret"

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
