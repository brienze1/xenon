package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FinishPullService interface {
	Finish(ctx context.Context, sessionId string) (*entity.SyncStatus, error)
}
