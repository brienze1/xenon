package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type update struct {
	errorHandler        adapter.ErrorHandler
	manageUpdateService appAdapter.ManageUpdateService
}

func NewUpdate(errorHandler adapter.ErrorHandler, manageUpdateService appAdapter.ManageUpdateService) adapter.UpdateController {
	return &update{errorHandler: errorHandler, manageUpdateService: manageUpdateService}
}

func (u *update) GetStatus(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.Status)
}

func (u *update) Check(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.Check)
}

func (u *update) CheckOnStartup(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.CheckOnStartup)
}

func (u *update) Update(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.Download)
}

func (u *update) Restart(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.Restart)
}

func (u *update) OpenReleasePage(bindingCtx context.Context) (*dto.UpdateStatus, error) {
	return u.run(bindingCtx, u.manageUpdateService.OpenReleasePage)
}

func (u *update) InstallOnQuit(bindingCtx context.Context) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			_ = u.errorHandler.Handle(ctx, handlePanic(ctx, nil, recovered), nil)
		}
	}(ctx)

	u.manageUpdateService.InstallOnQuit(ctx, currentInstallation())

	logger.Info(ctx, "Finished")
}

func (u *update) run(bindingCtx context.Context, action func(context.Context, entity.Installation) entity.UpdateStatus) (response *dto.UpdateStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, u.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	result := dto.NewUpdateStatus(action(ctx, currentInstallation()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func currentInstallation() entity.Installation {
	current := properties.Properties()
	return entity.Installation{
		Version:     current.Application.Version,
		Path:        current.Update.InstallPath,
		Platform:    entity.UpdatePlatform(current.Update.Platform),
		Installable: current.Update.Installable,
	}
}
