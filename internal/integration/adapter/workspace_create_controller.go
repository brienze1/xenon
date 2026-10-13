package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceCreateController interface {
	Handle(ctx context.Context, request dto.CreateWorkspace) (*dto.WorkspaceSummary, error)
}
