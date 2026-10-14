package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPTreeVariableLine struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Initial     string `json:"initial,omitempty"`
	Current     string `json:"current,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

// NewMCPTreeVariableLine masks a secret's initial value and drops its current one.
func NewMCPTreeVariableLine(row VariableRow) MCPTreeVariableLine {
	line := MCPTreeVariableLine{Key: row.Key, Description: row.Description, Initial: row.InitialValue}
	if row.HasCurrentOverride {
		line.Current = row.CurrentValue
	}
	if !row.Enabled {
		disabled := false
		line.Enabled = &disabled
	}
	if row.Secret {
		line.Secret, line.State, line.Initial, line.Current = true, row.SecretState, entity.MaskedValue, ""
	}
	return line
}

func NewMCPTreeVariableLines(rows []VariableRow) []MCPTreeVariableLine {
	lines := make([]MCPTreeVariableLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, NewMCPTreeVariableLine(row))
	}
	return lines
}
