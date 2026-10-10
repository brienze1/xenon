package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableEdit struct {
	SecretValue  *string `json:"secret_value,omitempty"`
	Enabled      *bool   `json:"enabled,omitempty"`
	Key          string  `json:"key"`
	PreviousKey  string  `json:"previous_key,omitempty"`
	InitialValue string  `json:"initial_value"`
	CurrentValue string  `json:"current_value,omitempty"`
	Description  string  `json:"description"`
	SetCurrent   bool    `json:"set_current"`
	Secret       bool    `json:"secret"`
}

func (v *VariableEdit) ToEntity() entity.VariableEdit {
	row := entity.NewVariableRow(v.Key, v.InitialValue)
	row.Secret = v.Secret
	row.Description = v.Description
	if v.Enabled != nil {
		row.Enabled = *v.Enabled
	}
	return entity.VariableEdit{Row: row, PreviousKey: v.PreviousKey, CurrentValue: v.CurrentValue, SetCurrent: v.SetCurrent, SecretValue: v.SecretValue}
}

func (v VariableEdit) LogValue() any {
	return map[string]any{"key": v.Key, "previous_key": v.PreviousKey, "secret": v.Secret, "set_current": v.SetCurrent, "sets_secret": v.SecretValue != nil}
}
