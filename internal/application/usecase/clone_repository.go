package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CloneRepository interface {
	Clone(ctx context.Context, access entity.GitAccess, clonePath string, settings entity.SyncSettings) error
}
