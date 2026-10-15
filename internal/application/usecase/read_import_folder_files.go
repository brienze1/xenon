package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadImportFolderFiles interface {
	ReadFolderFiles(ctx context.Context, folder entity.ImportInput) ([]entity.ImportInput, []string, []entity.ImportSkippedEntry, error)
}
