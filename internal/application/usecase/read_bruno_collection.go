package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadBrunoCollection interface {
	ReadCollection(ctx context.Context, input entity.ImportInput) (*entity.BrunoCollection, error)
}
