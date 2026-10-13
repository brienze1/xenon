package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceRemove struct {
	errorHandler           adapter.ErrorHandler
	removeWorkspaceService appAdapter.RemoveWorkspaceService
}

func NewWorkspaceRemove(errorHandler adapter.ErrorHandler, removeWorkspaceService appAdapter.RemoveWorkspaceService) adapter.WorkspaceRemoveController {
	return &workspaceRemove{errorHandler: errorHandler, removeWorkspaceService: removeWorkspaceService}
}

func (w *workspaceRemove) Handle(bindingCtx context.Context, request dto.RemoveWorkspace) (response *dto.WorkspaceList, err error) {
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

	err = validator.CustomValidator().Struct("WKS-23400", &request)
	if err != nil {
		return nil, err
	}

	registry, err := w.removeWorkspaceService.Remove(ctx, request.LocalId, request.MoveToTrash)
	if err != nil {
		return nil, err
	}
	result := dto.NewWorkspaceList(*registry)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
