package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CommitChangesService interface {
	Commit(ctx context.Context, localId, message string, itemIds []string) (*entity.SyncStatus, error)
}
