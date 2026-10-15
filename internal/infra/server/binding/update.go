package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type UpdateService struct {
	update adapter.UpdateController
}

func NewUpdateService(update adapter.UpdateController) *UpdateService {
	return &UpdateService{update: update}
}

func (u *UpdateService) GetStatus(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.GetStatus(ctx)
}

func (u *UpdateService) Check(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.Check(ctx)
}

func (u *UpdateService) CheckOnStartup(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.CheckOnStartup(ctx)
}

func (u *UpdateService) Update(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.Update(ctx)
}

func (u *UpdateService) Restart(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.Restart(ctx)
}

func (u *UpdateService) OpenReleasePage(ctx context.Context) (*dto.UpdateStatus, error) {
	return u.update.OpenReleasePage(ctx)
}

func (u *UpdateService) ServiceShutdown() error {
	u.update.InstallOnQuit(context.Background())
	return nil
}
