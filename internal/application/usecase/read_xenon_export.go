package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadXenonExport interface {
	ReadXenonExport(ctx context.Context, input entity.ImportInput) (*entity.XenonImport, error)
}
