package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PublishVariableEvent interface {
	VariablesChanged(ctx context.Context, localId string, scopes []entity.ScopeRef, reason string) error
	VaultStatusChanged(ctx context.Context, status entity.VaultStatus) error
}
