package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

func vaultFor(ctx context.Context, findVault usecase.FindVault, localId string, rows ...[]entity.VariableRow) (*entity.Vault, error) {
	if entity.AnyHasSecretRows(rows...) {
		return findVault.Find(ctx, localId)
	}
	vault := entity.NewUnopenedVault(localId)
	return &vault, nil
}
