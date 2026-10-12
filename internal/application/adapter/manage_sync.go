package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageSyncService interface {
	Status(ctx context.Context, localId string) (*entity.SyncStatus, error)
	Changes(ctx context.Context, localId string) ([]entity.ChangeItem, error)
	SaveSettings(ctx context.Context, localId, accountId string, autoPull, autoPush bool) error
}
