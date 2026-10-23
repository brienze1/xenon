package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type ExportRunService interface {
	Export(ctx context.Context, localId, runId string, format enums.RunExportFormat, path string) error
}
