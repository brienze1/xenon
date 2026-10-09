package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type treeSettings struct {
	errorHandler              adapter.ErrorHandler
	manageItemSettingsService appAdapter.ManageItemSettingsService
}

func NewTreeSettings(errorHandler adapter.ErrorHandler, manageItemSettingsService appAdapter.ManageItemSettingsService) adapter.TreeSettingsController {
	return &treeSettings{errorHandler: errorHandler, manageItemSettingsService: manageItemSettingsService}
}

func (t *treeSettings) GetSettings(bindingCtx context.Context, request dto.TreeItemReference) (response *dto.ItemSettings, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("TRE-12400", &request)
	if err != nil {
		return nil, err
	}

	settings, err := t.manageItemSettingsService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewItemSettings(*settings)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeSettings) SaveSettings(bindingCtx context.Context, request dto.SaveItemSettings) (response *dto.ItemSettings, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("TRE-13400", &request)
	if err != nil {
		return nil, err
	}

	settings, err := t.manageItemSettingsService.Save(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewItemSettings(*settings)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
