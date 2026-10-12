package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type syncConflict struct {
	errorHandler            adapter.ErrorHandler
	resolveConflictsService appAdapter.ResolveConflictsService
	finishPullService       appAdapter.FinishPullService
}

func NewSyncConflict(
	errorHandler adapter.ErrorHandler,
	resolveConflictsService appAdapter.ResolveConflictsService,
	finishPullService appAdapter.FinishPullService,
) adapter.SyncConflictController {
	return &syncConflict{errorHandler: errorHandler, resolveConflictsService: resolveConflictsService, finishPullService: finishPullService}
}

func (s *syncConflict) Session(bindingCtx context.Context, request dto.SyncWorkspaceReference) (response *dto.SyncSession, err error) {
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

	err = validator.CustomValidator().Struct("GIT-33400", &request)
	if err != nil {
		return nil, err
	}

	session, err := s.resolveConflictsService.Session(ctx, request.LocalId)
	if err != nil || session == nil {
		return nil, err
	}
	result := dto.NewSyncSession(*session)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncConflict) Resolve(bindingCtx context.Context, request dto.ResolveSyncConflict) (response *dto.SyncConflictItem, err error) {
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

	err = validator.CustomValidator().Struct("GIT-29400", &request)
	if err != nil {
		return nil, err
	}

	conflict, err := s.resolveConflictsService.Resolve(ctx, request.SessionId, request.ItemId, request.ConflictResolution(), request.FieldPath)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncConflictItem(*conflict)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncConflict) FinishPull(bindingCtx context.Context, request dto.SyncSessionReference) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-31400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.finishPullService.Finish(ctx, request.SessionId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncConflict) CancelPull(bindingCtx context.Context, request dto.SyncSessionReference) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-32400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.resolveConflictsService.Cancel(ctx, request.SessionId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
