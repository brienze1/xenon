package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type environmentManage struct {
	errorHandler              adapter.ErrorHandler
	manageEnvironmentService  appAdapter.ManageEnvironmentService
	reorderEnvironmentService appAdapter.ReorderEnvironmentService
	autoSyncService           appAdapter.AutoSyncService
}

func NewEnvironmentManage(
	errorHandler adapter.ErrorHandler,
	manageEnvironmentService appAdapter.ManageEnvironmentService,
	reorderEnvironmentService appAdapter.ReorderEnvironmentService,
	autoSyncService appAdapter.AutoSyncService,
) adapter.EnvironmentManageController {
	return &environmentManage{
		errorHandler:              errorHandler,
		manageEnvironmentService:  manageEnvironmentService,
		reorderEnvironmentService: reorderEnvironmentService,
		autoSyncService:           autoSyncService,
	}
}

func (e *environmentManage) Get(bindingCtx context.Context, request dto.EnvironmentReference) (response *dto.Environment, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateGetEnvironment(&request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageEnvironmentService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewEnvironment(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *environmentManage) Create(bindingCtx context.Context, request dto.CreateEnvironment) (response *dto.Environment, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateCreateEnvironment(&request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageEnvironmentService.Create(ctx, request.LocalId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewEnvironment(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *environmentManage) Update(bindingCtx context.Context, request dto.UpdateEnvironment) (response *dto.Environment, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateUpdateEnvironment(&request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageEnvironmentService.Update(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return nil, err
	}
	triggerAutoSync(ctx, e.autoSyncService.Saved, request.LocalId)
	result := dto.NewEnvironment(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *environmentManage) Duplicate(bindingCtx context.Context, request dto.EnvironmentReference) (response *dto.Environment, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateDuplicateEnvironment(&request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageEnvironmentService.Duplicate(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewEnvironment(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *environmentManage) Delete(bindingCtx context.Context, request dto.EnvironmentReference) (response *dto.ActiveEnvironment, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateDeleteEnvironment(&request)
	if err != nil {
		return nil, err
	}

	activeId, err := e.manageEnvironmentService.Delete(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewActiveEnvironment(activeId)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *environmentManage) Reorder(bindingCtx context.Context, request dto.ReorderEnvironment) (response *dto.EnvironmentList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("ENV-17400", &request)
	if err != nil {
		return nil, err
	}

	list, err := e.reorderEnvironmentService.Reorder(ctx, request.LocalId, request.Id, request.Index)
	if err != nil {
		return nil, err
	}
	triggerAutoSync(ctx, e.autoSyncService.Saved, request.LocalId)
	result := dto.NewEnvironmentList(*list)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
