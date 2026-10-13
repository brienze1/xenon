package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceList struct {
	errorHandler          adapter.ErrorHandler
	listWorkspacesService appAdapter.ListWorkspacesService
}

func NewWorkspaceList(
	errorHandler adapter.ErrorHandler,
	listWorkspacesService appAdapter.ListWorkspacesService,
) adapter.WorkspaceListController {
	return &workspaceList{
		errorHandler:          errorHandler,
		listWorkspacesService: listWorkspacesService,
	}
}

func (w *workspaceList) Handle(bindingCtx context.Context) (response *dto.WorkspaceList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, w.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	registry, err := w.listWorkspacesService.List(ctx)
	if err != nil {
		return nil, err
	}

	result := dto.NewWorkspaceList(*registry)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
