package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

func vaultFor(ctx context.Context, findVault usecase.FindVault, localId string, rows ...[]entity.VariableRow) (*entity.Vault, error) {
	for _, scopeRows := range rows {
		if entity.HasSecretRows(scopeRows) {
			return findVault.Find(ctx, localId)
		}
	}
	return &entity.Vault{LocalID: localId}, nil
}
