package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SyncLinkController interface {
	Link(ctx context.Context, request dto.LinkRepository) (*dto.SyncStatus, error)
	CloneAsWorkspace(ctx context.Context, request dto.CloneWorkspace) (*dto.WorkspaceSummary, error)
}
