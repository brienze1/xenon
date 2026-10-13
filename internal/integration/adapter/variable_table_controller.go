package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type VariableTableController interface {
	GetTable(ctx context.Context, request dto.VariableTableReference) (*dto.VariableTable, error)
	SaveTable(ctx context.Context, request dto.SaveVariableTable) (*dto.VariableTable, error)
	SetCurrent(ctx context.Context, request dto.SetCurrentValue) (*dto.VariableRow, error)
	Reset(ctx context.Context, request dto.VariableKeys) (*dto.VariableTable, error)
	Persist(ctx context.Context, request dto.VariableKeys) (*dto.VariableTable, error)
	AddVariable(ctx context.Context, request dto.AddVariable) (*dto.VariableRow, error)
}
