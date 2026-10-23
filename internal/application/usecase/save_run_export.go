package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type SaveRunExport interface {
	SaveRunExport(ctx context.Context, path string, run entity.Run, format enums.RunExportFormat) error
}
