package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindDigestState interface {
	FindDigestState(ctx context.Context, localId, owner string) (*entity.DigestState, error)
}
