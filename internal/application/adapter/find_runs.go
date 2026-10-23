package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindRunsService interface {
	Get(ctx context.Context, localId, runId string) (*entity.Run, error)
	List(ctx context.Context, localId, itemId string) ([]entity.Run, error)
	// ListForAI resolves the target by path or item id; the target comes back with the runs and with any error.
	ListForAI(ctx context.Context, localId, path, itemId string) ([]entity.Run, entity.AIActivityTarget, error)
	Active(ctx context.Context, localId string) ([]entity.Run, error)
	Response(ctx context.Context, localId, runId string, index int) (*entity.Response, error)
}
