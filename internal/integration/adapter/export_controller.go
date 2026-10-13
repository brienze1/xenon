package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ExportController interface {
	Options(ctx context.Context, request dto.ExportItemReference) (*dto.ExportOptions, error)
	PickTarget(ctx context.Context, request dto.PickExportTarget) (*dto.PickedPath, error)
	Export(ctx context.Context, request dto.Export) (*dto.ExportSummary, error)
	ShowInFolder(ctx context.Context, request dto.ShowInFolder) error
}
