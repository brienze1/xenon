package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveVaultSettings interface {
	SaveSettings(ctx context.Context, settings entity.VaultSettings) error
}
