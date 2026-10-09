package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type uiState struct {
	errorHandler         adapter.ErrorHandler
	manageUIStateService appAdapter.ManageUIStateService
}

func NewUIState(errorHandler adapter.ErrorHandler, manageUIStateService appAdapter.ManageUIStateService) adapter.UIStateController {
	return &uiState{errorHandler: errorHandler, manageUIStateService: manageUIStateService}
}

func (u *uiState) Get(bindingCtx context.Context, request dto.GetUIState) (response *dto.UIState, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, u.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("UIS-00400", &request)
	if err != nil {
		return nil, err
	}

	state, err := u.manageUIStateService.Get(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewUIState(request.LocalId, *state)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (u *uiState) Save(bindingCtx context.Context, request dto.UIState) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = u.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("UIS-01400", &request)
	if err != nil {
		return err
	}

	err = u.manageUIStateService.Save(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
