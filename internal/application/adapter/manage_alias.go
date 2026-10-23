package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageAliasService interface {
	Create(ctx context.Context, localId, requestId, name string, environmentIds []string, baseRevision string, force bool) (*entity.Request, error)
	Save(ctx context.Context, localId, requestId string, alias entity.RequestAlias, baseRevision string, force bool) (*entity.Request, error)
	SetEnvironments(ctx context.Context, localId, requestId, aliasId string, environmentIds []string, baseRevision string, force bool) (*entity.Request, error)
	Delete(ctx context.Context, localId, requestId, aliasId, baseRevision string, force bool) (*entity.Request, error)
}
