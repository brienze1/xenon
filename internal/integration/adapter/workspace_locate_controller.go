package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceLocateController interface {
	Handle(ctx context.Context, request dto.LocateWorkspace) (*dto.WorkspaceSummary, error)
}
