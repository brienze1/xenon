package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveVault interface {
	Save(ctx context.Context, vault entity.Vault) error
}
