package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceListController interface {
	Handle(ctx context.Context) (*dto.WorkspaceList, error)
}
