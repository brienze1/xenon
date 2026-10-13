package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RollbackImportItem interface {
	Rollback(ctx context.Context, staged entity.StagedImport) error
}
