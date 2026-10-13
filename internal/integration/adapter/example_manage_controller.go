package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ExampleManageController interface {
	Get(ctx context.Context, request dto.ExampleReference) (*dto.Example, error)
	Create(ctx context.Context, request dto.CreateExample) (*dto.Example, error)
	CreateFromResponse(ctx context.Context, request dto.CreateExampleFromResponse) (*dto.Example, error)
	Save(ctx context.Context, request dto.SaveExample) (*dto.Example, error)
	Rename(ctx context.Context, request dto.RenameExample) (*dto.Example, error)
	Duplicate(ctx context.Context, request dto.ExampleReference) (*dto.Example, error)
	Delete(ctx context.Context, request dto.ExampleReference) error
}
