package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type LinkWorkspaceService interface {
	Link(ctx context.Context, localId string, settings entity.SyncSettings, accountId, clonePath string, mode enums.LinkMode) (*entity.SyncStatus, error)
}
