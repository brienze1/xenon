package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindIterationData interface {
	FindIterationData(ctx context.Context, path string) (*entity.IterationData, error)
	IterationDataExists(ctx context.Context, path string) bool
}
