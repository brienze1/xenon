package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveDigestState interface {
	SaveDigestState(ctx context.Context, localId, owner string, state entity.DigestState) error
}
