package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ImportManageController interface {
	Detect(ctx context.Context, request dto.DetectImport) (*dto.ImportPreview, error)
	PickSources(ctx context.Context) (*dto.PickedSources, error)
	Clashes(ctx context.Context, request dto.ImportPlan) (*dto.ImportClashes, error)
	Import(ctx context.Context, request dto.ImportPlan) (*dto.ImportResult, error)
	Discard(ctx context.Context, request dto.ImportSourceReference) error
}
