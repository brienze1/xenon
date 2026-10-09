package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceRenameController interface {
	Handle(ctx context.Context, request dto.RenameWorkspace) (*dto.WorkspaceSummary, error)
}
