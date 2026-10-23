package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type aliasManage struct {
	errorHandler       adapter.ErrorHandler
	manageAliasService appAdapter.ManageAliasService
}

func NewAliasManage(errorHandler adapter.ErrorHandler, manageAliasService appAdapter.ManageAliasService) adapter.AliasManageController {
	return &aliasManage{errorHandler: errorHandler, manageAliasService: manageAliasService}
}

func (a *aliasManage) Create(bindingCtx context.Context, request dto.CreateAlias) (response *dto.Request, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateCreateAlias(&request)
	if err != nil {
		return nil, err
	}

	saved, err := a.manageAliasService.Create(ctx, request.LocalId, request.RequestId, request.Name, request.EnvironmentIds, request.BaseRevision, false)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aliasManage) Save(bindingCtx context.Context, request dto.SaveAlias) (response *dto.Request, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateSaveAlias(&request)
	if err != nil {
		return nil, err
	}

	saved, err := a.manageAliasService.Save(ctx, request.LocalId, request.RequestId, request.ToEntity(), request.BaseRevision, request.Force)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aliasManage) SetEnvironments(bindingCtx context.Context, request dto.SetAliasEnvironments) (response *dto.Request, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateSetAliasEnvironments(&request)
	if err != nil {
		return nil, err
	}

	saved, err := a.manageAliasService.SetEnvironments(ctx, request.LocalId, request.RequestId, request.AliasId, request.EnvironmentIds, request.BaseRevision, false)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aliasManage) Delete(bindingCtx context.Context, request dto.AliasReference) (response *dto.Request, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateDeleteAlias(&request)
	if err != nil {
		return nil, err
	}

	saved, err := a.manageAliasService.Delete(ctx, request.LocalId, request.RequestId, request.AliasId, request.BaseRevision, false)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
