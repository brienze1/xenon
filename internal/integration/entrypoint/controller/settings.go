package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type settings struct {
	errorHandler             adapter.ErrorHandler
	manageAppSettingsService appAdapter.ManageAppSettingsService
}

func NewSettings(errorHandler adapter.ErrorHandler, manageAppSettingsService appAdapter.ManageAppSettingsService) adapter.SettingsController {
	return &settings{errorHandler: errorHandler, manageAppSettingsService: manageAppSettingsService}
}

func (s *settings) Get(bindingCtx context.Context) (response *dto.AppSettings, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, s.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	found, err := s.manageAppSettingsService.Get(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewAppSettings(*found, properties.Properties().Application.MCPPortOverride)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *settings) Save(bindingCtx context.Context, request dto.AppSettings) (response *dto.AppSettings, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, s.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("SET-00400", &request)
	if err != nil {
		return nil, err
	}

	saved, err := s.manageAppSettingsService.Save(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewAppSettings(*saved, properties.Properties().Application.MCPPortOverride)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
