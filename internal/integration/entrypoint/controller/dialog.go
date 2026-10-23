package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type dialog struct {
	errorHandler    adapter.ErrorHandler
	pickPathService appAdapter.PickPathService
}

func NewDialog(errorHandler adapter.ErrorHandler, pickPathService appAdapter.PickPathService) adapter.DialogController {
	return &dialog{errorHandler: errorHandler, pickPathService: pickPathService}
}

func (d *dialog) PickFolder(bindingCtx context.Context, request dto.PickFolder) (response *dto.PickedPath, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, d.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DLG-03400", &request)
	if err != nil {
		return nil, err
	}

	path, err := d.pickPathService.PickFolder(ctx, request.Title)
	if err != nil {
		return nil, err
	}
	result := dto.NewPickedPath(path)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (d *dialog) PickFile(bindingCtx context.Context, request dto.PickFile) (response *dto.PickedPath, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, d.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DLG-04400", &request)
	if err != nil {
		return nil, err
	}

	path, err := d.pickPathService.PickFile(ctx, request.Title, request.FileFilters())
	if err != nil {
		return nil, err
	}
	result := dto.NewPickedPath(path)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (d *dialog) PickSaveFile(bindingCtx context.Context, request dto.PickSaveFile) (response *dto.PickedPath, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, d.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DLG-05400", &request)
	if err != nil {
		return nil, err
	}

	path, err := d.pickPathService.PickSaveFile(ctx, request.Title, request.SuggestedName)
	if err != nil {
		return nil, err
	}
	result := dto.NewPickedPath(path)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
