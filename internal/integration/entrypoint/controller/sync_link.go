package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type syncLink struct {
	errorHandler           adapter.ErrorHandler
	linkWorkspaceService   appAdapter.LinkWorkspaceService
	cloneWorkspaceService  appAdapter.CloneWorkspaceService
	unlinkWorkspaceService appAdapter.UnlinkWorkspaceService
}

func NewSyncLink(
	errorHandler adapter.ErrorHandler,
	linkWorkspaceService appAdapter.LinkWorkspaceService,
	cloneWorkspaceService appAdapter.CloneWorkspaceService,
	unlinkWorkspaceService appAdapter.UnlinkWorkspaceService,
) adapter.SyncLinkController {
	return &syncLink{errorHandler: errorHandler, linkWorkspaceService: linkWorkspaceService, cloneWorkspaceService: cloneWorkspaceService, unlinkWorkspaceService: unlinkWorkspaceService}
}

func (s *syncLink) Link(bindingCtx context.Context, request dto.LinkRepository) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-63400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.linkWorkspaceService.Link(ctx, request.LocalId, request.ToEntity(), request.AccountID, request.ClonePath, request.LinkMode())
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncLink) CloneAsWorkspace(bindingCtx context.Context, request dto.CloneWorkspace) (response *dto.WorkspaceSummary, err error) {
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

	err = validator.CustomValidator().Struct("GIT-39400", &request)
	if err != nil {
		return nil, err
	}

	registered, err := s.cloneWorkspaceService.Clone(ctx, request.ToEntity(), request.AccountID, request.ClonePath, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewWorkspaceSummary(*registered)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncLink) Unlink(bindingCtx context.Context, request dto.UnlinkRepository) (response *dto.WorkspaceSummary, err error) {
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

	err = validator.CustomValidator().Struct("GIT-40400", &request)
	if err != nil {
		return nil, err
	}

	registered, err := s.unlinkWorkspaceService.Unlink(ctx, request.LocalId, request.GetTargetFolder())
	if err != nil {
		return nil, err
	}
	result := dto.NewWorkspaceSummary(*registered)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
