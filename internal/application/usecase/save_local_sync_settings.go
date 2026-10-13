package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveLocalSyncSettings interface {
	SaveSyncSettings(ctx context.Context, localId string, settings entity.LocalSyncSettings) error
}
