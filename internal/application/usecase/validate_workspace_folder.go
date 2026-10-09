package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ValidateWorkspaceFolder interface {
	ValidateNew(ctx context.Context, root string, registry entity.AppRegistry) error
	ValidateExisting(ctx context.Context, root, localId string, registry entity.AppRegistry) error
}
