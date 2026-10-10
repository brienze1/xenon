package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageVariableTableService interface {
	GetTable(ctx context.Context, localId string, scope entity.ScopeRef) (*entity.VariableTable, error)
	SaveTable(ctx context.Context, localId string, scope entity.ScopeRef, edits []entity.VariableEdit, confirmReveal bool) (*entity.VariableTable, error)
	SetCurrent(ctx context.Context, localId string, scope entity.ScopeRef, key, value string) (*entity.VariableTableEntry, error)
	Reset(ctx context.Context, localId string, scope entity.ScopeRef, keys []string) (*entity.VariableTable, error)
	Persist(ctx context.Context, localId string, scope entity.ScopeRef, keys []string) (*entity.VariableTable, error)
	AddVariable(ctx context.Context, localId string, scope entity.ScopeRef, row entity.VariableRow, value string) (*entity.VariableTableEntry, error)
}
