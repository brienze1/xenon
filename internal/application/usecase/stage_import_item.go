package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type StageImportItem interface {
	Stage(ctx context.Context, staged entity.StagedImport) (*entity.StagedImport, error)
}
