package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAppSettings interface {
	FindAppSettings(ctx context.Context) (*entity.AppSettings, error)
}
