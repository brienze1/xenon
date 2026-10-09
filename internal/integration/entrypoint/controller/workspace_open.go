package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceOpen struct {
	errorHandler         adapter.ErrorHandler
	openWorkspaceService appAdapter.OpenWorkspaceService
}

func NewWorkspaceOpen(errorHandler adapter.ErrorHandler, openWorkspaceService appAdapter.OpenWorkspaceService) adapter.WorkspaceOpenController {
	return &workspaceOpen{errorHandler: errorHandler, openWorkspaceService: openWorkspaceService}
}

func (w *workspaceOpen) Handle(bindingCtx context.Context, request dto.OpenWorkspace) (response *dto.WorkspaceSummary, err error) {
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

	err = validator.CustomValidator().Struct("WKS-21400", &request)
	if err != nil {
		return nil, err
	}

	workspace, err := w.openWorkspaceService.Open(ctx, request.GetPath())
	if err != nil {
		return nil, err
	}
	result := dto.NewWorkspaceSummary(*workspace)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
