package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableRow struct {
	Key                string `json:"key"`
	InitialValue       string `json:"initial_value"`
	CurrentValue       string `json:"current_value"`
	SecretState        string `json:"secret_state,omitempty"`
	Description        string `json:"description"`
	HasCurrentOverride bool   `json:"has_current_override"`
	Secret             bool   `json:"secret"`
	Enabled            bool   `json:"enabled"`
	Duplicate          bool   `json:"duplicate"`
	SurroundingSpaces  bool   `json:"surrounding_spaces"`
}

func NewVariableRow(entry entity.VariableTableEntry) VariableRow {
	row := VariableRow{
		Key:                entry.Row.Key,
		Description:        entry.Row.Description,
		HasCurrentOverride: entry.HasCurrentOverride,
		Secret:             entry.Row.Secret,
		Enabled:            entry.Row.Enabled,
		Duplicate:          entry.Duplicate,
		SurroundingSpaces:  entry.HasSurroundingSpaces(),
	}
	if entry.Row.Secret {
		row.SecretState = entry.SecretState.String()
		return row
	}
	row.InitialValue = entry.Row.InitialValue
	row.CurrentValue = entry.CurrentValue
	return row
}

func NewVariableRows(entries []entity.VariableTableEntry) []VariableRow {
	rows := make([]VariableRow, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, NewVariableRow(entry))
	}
	return rows
}

func (v VariableRow) LogValue() any {
	return map[string]any{"key": v.Key, "secret": v.Secret, "secret_state": v.SecretState, "enabled": v.Enabled}
}
