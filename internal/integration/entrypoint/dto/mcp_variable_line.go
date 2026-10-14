package dto

type MCPVariableLine struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Initial     string `json:"initial,omitempty"`
	Current     string `json:"current,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

// NewMCPVariableLine never carries a secret row's values, only its state.
func NewMCPVariableLine(row VariableRow) MCPVariableLine {
	line := MCPVariableLine{Key: row.Key, Description: row.Description}
	if !row.Enabled {
		disabled := false
		line.Enabled = &disabled
	}
	if row.Secret {
		line.Secret, line.State = true, row.SecretState
		return line
	}
	line.Initial = row.InitialValue
	if row.HasCurrentOverride {
		line.Current = row.CurrentValue
	}
	return line
}

func NewMCPVariableLines(rows []VariableRow) []MCPVariableLine {
	lines := make([]MCPVariableLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, NewMCPVariableLine(row))
	}
	return lines
}
