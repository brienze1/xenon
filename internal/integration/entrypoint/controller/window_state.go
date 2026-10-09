package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type windowState struct {
	errorHandler       adapter.ErrorHandler
	windowStateService appAdapter.WindowStateService
}

func NewWindowState(errorHandler adapter.ErrorHandler, windowStateService appAdapter.WindowStateService) adapter.WindowStateController {
	return &windowState{errorHandler: errorHandler, windowStateService: windowStateService}
}

func (w *windowState) Restore(bindingCtx context.Context, request dto.RestoreWindow) (response *dto.WindowState, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, w.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("WIN-00400", &request)
	if err != nil {
		return nil, err
	}

	state, err := w.windowStateService.Restore(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewWindowState(*state)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (w *windowState) Save(bindingCtx context.Context, request dto.WindowState) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = w.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("WIN-01400", &request)
	if err != nil {
		return err
	}

	err = w.windowStateService.Save(ctx, request.ToEntity())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
