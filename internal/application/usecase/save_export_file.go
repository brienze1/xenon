package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveExportFile interface {
	SaveFile(ctx context.Context, path string, data []byte) error
	SaveArchive(ctx context.Context, path string, entries []entity.ExportEntry) error
}
