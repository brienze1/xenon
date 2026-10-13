package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindImportSource interface {
	Find(ctx context.Context, id string) (*entity.ImportSource, error)
}
