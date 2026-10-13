package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRequestService interface {
	Open(ctx context.Context, localId, id string) (*entity.Request, error)
	Save(ctx context.Context, localId string, request entity.Request, baseRevision string, force bool) (*entity.Request, error)
	SaveAs(ctx context.Context, localId, parentId, name string, request entity.Request) (*entity.Request, error)
}
