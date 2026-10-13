package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RemoveWorkspaceService interface {
	Remove(ctx context.Context, localId string, moveToTrash bool) (*entity.AppRegistry, error)
}
