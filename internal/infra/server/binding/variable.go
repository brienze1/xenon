package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type VariableService struct {
	table adapter.VariableTableController
}

func NewVariableService(table adapter.VariableTableController) *VariableService {
	return &VariableService{table: table}
}

func (v *VariableService) GetTable(ctx context.Context, request dto.VariableTableReference) (*dto.VariableTable, error) {
	return v.table.GetTable(ctx, request)
}

func (v *VariableService) SaveTable(ctx context.Context, request dto.SaveVariableTable) (*dto.VariableTable, error) {
	return v.table.SaveTable(ctx, request)
}

func (v *VariableService) SetCurrent(ctx context.Context, request dto.SetCurrentValue) (*dto.VariableRow, error) {
	return v.table.SetCurrent(ctx, request)
}

func (v *VariableService) Reset(ctx context.Context, request dto.VariableKeys) (*dto.VariableTable, error) {
	return v.table.Reset(ctx, request)
}

func (v *VariableService) Persist(ctx context.Context, request dto.VariableKeys) (*dto.VariableTable, error) {
	return v.table.Persist(ctx, request)
}

func (v *VariableService) AddVariable(ctx context.Context, request dto.AddVariable) (*dto.VariableRow, error) {
	return v.table.AddVariable(ctx, request)
}
