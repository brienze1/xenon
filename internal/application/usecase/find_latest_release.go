package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindLatestRelease interface {
	FindLatestRelease(ctx context.Context) (*entity.Release, error)
}
