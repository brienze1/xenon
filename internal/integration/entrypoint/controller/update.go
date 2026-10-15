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

func (u *update) GetStatus(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	result := dto.NewUpdateStatus(u.manageUpdateService.Status(ctx, currentInstallation()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) Check(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	result := dto.NewUpdateStatus(u.manageUpdateService.Check(ctx, currentInstallation()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) CheckOnStartup(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	status, err := u.manageUpdateService.CheckOnStartup(ctx, currentInstallation())
	if err != nil {
		return nil, err
	}
	result := dto.NewUpdateStatus(status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) Update(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	result := dto.NewUpdateStatus(u.manageUpdateService.Download(ctx, currentInstallation()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) Restart(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	result := dto.NewUpdateStatus(u.manageUpdateService.Restart(ctx, currentInstallation()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) OpenReleasePage(bindingCtx context.Context) (response *dto.UpdateStatus, err error) {
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

	status, err := u.manageUpdateService.OpenReleasePage(ctx, currentInstallation())
	if err != nil {
		return nil, err
	}
	result := dto.NewUpdateStatus(status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *update) InstallOnQuit(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = u.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	u.manageUpdateService.InstallOnQuit(ctx, currentInstallation())

	logger.Info(ctx, "Finished")
	return nil
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
