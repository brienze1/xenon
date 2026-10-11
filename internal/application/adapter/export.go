package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type ExportService interface {
	Options(ctx context.Context, localId string, kind enums.ExportItemKind, itemId string) (*entity.ExportOptions, error)
	PickTarget(ctx context.Context, suggestedFileName string, format enums.ExportFormat) (string, error)
	Export(ctx context.Context, request entity.ExportRequest) (*entity.ExportSummary, error)
	ShowInFolder(ctx context.Context, path string) error
}
