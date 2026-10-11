package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ExportService struct {
	export adapter.ExportController
}

func NewExportService(export adapter.ExportController) *ExportService {
	return &ExportService{export: export}
}

func (e *ExportService) Options(ctx context.Context, request dto.ExportItemReference) (*dto.ExportOptions, error) {
	return e.export.Options(ctx, request)
}

func (e *ExportService) PickTarget(ctx context.Context, request dto.PickExportTarget) (*dto.PickedPath, error) {
	return e.export.PickTarget(ctx, request)
}

func (e *ExportService) Export(ctx context.Context, request dto.Export) (*dto.ExportSummary, error) {
	return e.export.Export(ctx, request)
}

func (e *ExportService) ShowInFolder(ctx context.Context, request dto.ShowInFolder) error {
	return e.export.ShowInFolder(ctx, request)
}
