package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveUpdateSession interface {
	SaveUpdateSession(ctx context.Context, session entity.UpdateSession)
	// ChangeUpdateSession applies change atomically and stores its result only when change accepts it.
	ChangeUpdateSession(ctx context.Context, change func(entity.UpdateSession) (entity.UpdateSession, bool)) (entity.UpdateSession, bool)
}
