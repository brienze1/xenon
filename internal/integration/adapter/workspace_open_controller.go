package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceOpenController interface {
	Handle(ctx context.Context, request dto.OpenWorkspace) (*dto.WorkspaceSummary, error)
}
