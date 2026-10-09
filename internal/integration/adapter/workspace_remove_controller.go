package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceRemoveController interface {
	Handle(ctx context.Context, request dto.RemoveWorkspace) (*dto.WorkspaceList, error)
}
