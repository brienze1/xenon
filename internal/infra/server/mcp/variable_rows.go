package mcp

import (
	"context"
	"errors"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

const revealRefusedCode = "VAR-01409"

type variableLine struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Initial     string `json:"initial,omitempty"`
	Current     string `json:"current,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type variableRowArguments struct {
	SecretValue *string `json:"secret_value"`
	Enabled     *bool   `json:"enabled"`
	Key         string  `json:"key"`
	Initial     string  `json:"initial"`
	Description string  `json:"description"`
	Secret      bool    `json:"secret"`
}

// newVariableLine never carries a secret row's values, only its state.
func newVariableLine(row dto.VariableRow) variableLine {
	line := variableLine{Key: row.Key, Description: row.Description}
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

func newVariableLines(rows []dto.VariableRow) []variableLine {
	lines := make([]variableLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, newVariableLine(row))
	}
	return lines
}

func (v variableRowArguments) edit() dto.VariableEdit {
	edit := dto.VariableEdit{Key: v.Key, Description: v.Description, Secret: v.Secret, Enabled: v.Enabled, SecretValue: v.SecretValue}
	if !v.Secret {
		edit.InitialValue = v.Initial
	}
	return edit
}

// saveVariableRows never confirms a reveal, so a row turning a stored secret plain is refused (MCP-02403) unwritten.
func (t *tools) saveVariableRows(ctx context.Context, call Call, scope string, rows []variableRowArguments) (*dto.VariableTable, error) {
	edits := make([]dto.VariableEdit, 0, len(rows))
	for _, row := range rows {
		if row.SecretValue != nil && !row.Secret {
			return nil, invalidArguments(call.Tool, "secret_value is only accepted on a row with secret: true")
		}
		edits = append(edits, row.edit())
	}
	table, err := t.controllers.VariableTable.SaveTable(ctx, dto.SaveVariableTable{LocalId: call.LocalId, Scope: scope, Rows: edits})
	if failedWithCode(err, revealRefusedCode) {
		return nil, humanOnly("Turning a secret variable plain")
	}
	return table, err
}

func failedWithCode(err error, code string) bool {
	var custom errs.BaseError
	return errors.As(err, &custom) && custom.Code() == code
}
