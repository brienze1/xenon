package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type UpdateController interface {
	GetStatus(ctx context.Context) (*dto.UpdateStatus, error)
	Check(ctx context.Context) (*dto.UpdateStatus, error)
	CheckOnStartup(ctx context.Context) (*dto.UpdateStatus, error)
	Update(ctx context.Context) (*dto.UpdateStatus, error)
	Restart(ctx context.Context) (*dto.UpdateStatus, error)
	OpenReleasePage(ctx context.Context) (*dto.UpdateStatus, error)
	InstallOnQuit(ctx context.Context) error
}
