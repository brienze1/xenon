package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindLocalSyncSettings interface {
	FindSyncSettings(ctx context.Context, localId string) (*entity.LocalSyncSettings, error)
}
