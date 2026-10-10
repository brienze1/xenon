package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveAppSettings interface {
	SaveAppSettings(ctx context.Context, settings entity.AppSettings) error
}
