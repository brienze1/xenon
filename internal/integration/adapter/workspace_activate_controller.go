package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceActivateController interface {
	Handle(ctx context.Context, request dto.ActivateWorkspace) (*dto.WorkspaceSnapshot, error)
}
