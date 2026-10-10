package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ImportService struct {
	manage adapter.ImportManageController
}

func NewImportService(manage adapter.ImportManageController) *ImportService {
	return &ImportService{manage: manage}
}

func (i *ImportService) Detect(ctx context.Context, request dto.DetectImport) (*dto.ImportPreview, error) {
	return i.manage.Detect(ctx, request)
}

func (i *ImportService) PickSources(ctx context.Context) (*dto.PickedSources, error) {
	return i.manage.PickSources(ctx)
}

func (i *ImportService) Clashes(ctx context.Context, request dto.ImportPlan) (*dto.ImportClashes, error) {
	return i.manage.Clashes(ctx, request)
}

func (i *ImportService) Import(ctx context.Context, request dto.ImportPlan) (*dto.ImportResult, error) {
	return i.manage.Import(ctx, request)
}

func (i *ImportService) Discard(ctx context.Context, request dto.ImportSourceReference) error {
	return i.manage.Discard(ctx, request)
}
