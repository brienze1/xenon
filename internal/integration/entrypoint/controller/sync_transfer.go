package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type syncTransfer struct {
	errorHandler         adapter.ErrorHandler
	commitChangesService appAdapter.CommitChangesService
	pullWorkspaceService appAdapter.PullWorkspaceService
	pushWorkspaceService appAdapter.PushWorkspaceService
}

func NewSyncTransfer(
	errorHandler adapter.ErrorHandler,
	commitChangesService appAdapter.CommitChangesService,
	pullWorkspaceService appAdapter.PullWorkspaceService,
	pushWorkspaceService appAdapter.PushWorkspaceService,
) adapter.SyncTransferController {
	return &syncTransfer{
		errorHandler:         errorHandler,
		commitChangesService: commitChangesService,
		pullWorkspaceService: pullWorkspaceService,
		pushWorkspaceService: pushWorkspaceService,
	}
}

func (s *syncTransfer) Commit(bindingCtx context.Context, request dto.CommitSyncChanges) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-00400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.commitChangesService.Commit(ctx, request.LocalId, request.Message, request.ItemIDs)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncTransfer) Pull(bindingCtx context.Context, request dto.SyncWorkspaceReference) (response *dto.SyncPullResult, err error) {
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

	err = validator.CustomValidator().Struct("GIT-69400", &request)
	if err != nil {
		return nil, err
	}

	pulled, err := s.pullWorkspaceService.Pull(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncPullResult(*pulled)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncTransfer) Push(bindingCtx context.Context, request dto.SyncWorkspaceReference) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-70400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.pushWorkspaceService.Push(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
