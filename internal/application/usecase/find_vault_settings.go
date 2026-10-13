package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindVaultSettings interface {
	FindSettings(ctx context.Context) (*entity.VaultSettings, error)
}
