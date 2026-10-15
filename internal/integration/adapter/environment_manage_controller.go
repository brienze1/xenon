package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type EnvironmentManageController interface {
	Get(ctx context.Context, request dto.EnvironmentReference) (*dto.Environment, error)
	Create(ctx context.Context, request dto.CreateEnvironment) (*dto.Environment, error)
	Update(ctx context.Context, request dto.UpdateEnvironment) (*dto.Environment, error)
	Duplicate(ctx context.Context, request dto.EnvironmentReference) (*dto.Environment, error)
	Delete(ctx context.Context, request dto.EnvironmentReference) (*dto.ActiveEnvironment, error)
	Reorder(ctx context.Context, request dto.ReorderEnvironment) (*dto.EnvironmentList, error)
}
