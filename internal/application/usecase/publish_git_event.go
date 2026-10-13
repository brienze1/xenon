package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishGitEvent interface {
	HostKeyUnknown(ctx context.Context, repository string, key entity.HostKey) error
	StatusChanged(ctx context.Context, localId string, status entity.SyncStatus) error
	Progress(ctx context.Context, localId, phase string, done, total int) error
}
