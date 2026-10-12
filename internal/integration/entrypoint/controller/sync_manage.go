package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type syncManage struct {
	errorHandler      adapter.ErrorHandler
	manageSyncService appAdapter.ManageSyncService
}

func NewSyncManage(errorHandler adapter.ErrorHandler, manageSyncService appAdapter.ManageSyncService) adapter.SyncManageController {
	return &syncManage{errorHandler: errorHandler, manageSyncService: manageSyncService}
}

func (s *syncManage) Status(bindingCtx context.Context, request dto.SyncWorkspaceReference) (response *dto.SyncStatus, err error) {
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

	err = validator.CustomValidator().Struct("GIT-67400", &request)
	if err != nil {
		return nil, err
	}

	status, err := s.manageSyncService.Status(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *syncManage) Changes(bindingCtx context.Context, request dto.SyncWorkspaceReference) (response []dto.SyncChangeItem, err error) {
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

	err = validator.CustomValidator().Struct("GIT-68400", &request)
	if err != nil {
		return nil, err
	}

	changes, err := s.manageSyncService.Changes(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSyncChangeItems(changes)

	logger.Info(ctx, "Finished", map[string]any{"changes": len(result)})
	return result, nil
}

func (s *syncManage) SaveSettings(bindingCtx context.Context, request dto.SaveSyncSettings) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = s.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("GIT-71400", &request)
	if err != nil {
		return err
	}

	err = s.manageSyncService.SaveSettings(ctx, request.LocalId, request.AccountID, request.AutoPull, request.AutoPush)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
