package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceActivate struct {
	errorHandler             adapter.ErrorHandler
	activateWorkspaceService appAdapter.ActivateWorkspaceService
}

func NewWorkspaceActivate(
	errorHandler adapter.ErrorHandler,
	activateWorkspaceService appAdapter.ActivateWorkspaceService,
) adapter.WorkspaceActivateController {
	return &workspaceActivate{
		errorHandler:             errorHandler,
		activateWorkspaceService: activateWorkspaceService,
	}
}

func (w *workspaceActivate) Handle(bindingCtx context.Context, request dto.ActivateWorkspace) (response *dto.WorkspaceSnapshot, err error) {
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

	err = validator.CustomValidator().Struct("WKS-08400", &request)
	if err != nil {
		return nil, err
	}

	snapshot, err := w.activateWorkspaceService.Activate(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}

	result := dto.NewWorkspaceSnapshot(*snapshot)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
