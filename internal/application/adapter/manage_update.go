package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageUpdateService interface {
	Status(ctx context.Context, installation entity.Installation) entity.UpdateStatus
	Check(ctx context.Context, installation entity.Installation) entity.UpdateStatus
	CheckOnStartup(ctx context.Context, installation entity.Installation) (entity.UpdateStatus, error)
	Download(ctx context.Context, installation entity.Installation) entity.UpdateStatus
	Restart(ctx context.Context, installation entity.Installation) entity.UpdateStatus
	InstallOnQuit(ctx context.Context, installation entity.Installation)
	OpenReleasePage(ctx context.Context, installation entity.Installation) (entity.UpdateStatus, error)
}
