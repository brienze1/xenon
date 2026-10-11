package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type WriteXenonExport interface {
	WriteXenonExport(ctx context.Context, root string, export entity.XenonExport) ([]entity.ExportEntry, error)
}
