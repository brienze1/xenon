package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishImportProgress interface {
	Progress(ctx context.Context, progress entity.ImportProgress) error
}
