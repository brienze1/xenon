package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindReleaseChecksums interface {
	FindReleaseChecksums(ctx context.Context, asset entity.ReleaseAsset) (entity.ReleaseChecksums, error)
}
