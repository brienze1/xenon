package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CommitImportItem interface {
	Commit(ctx context.Context, staged entity.StagedImport) error
}
