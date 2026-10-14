package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadPostmanCollection interface {
	ReadCollection(ctx context.Context, document entity.ImportDocument) (*entity.ImportCollection, error)
}
