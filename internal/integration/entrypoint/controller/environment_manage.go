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
	errorHandler             adapter.ErrorHandler
	manageEnvironmentService appAdapter.ManageEnvironmentService
}

func NewEnvironmentManage(errorHandler adapter.ErrorHandler, manageEnvironmentService appAdapter.ManageEnvironmentService) adapter.EnvironmentManageController {
	return &environmentManage{errorHandler: errorHandler, manageEnvironmentService: manageEnvironmentService}
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

	err = validator.CustomValidator().Struct("ENV-13400", &request)
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

	err = validator.CustomValidator().Struct("ENV-11400", &request)
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

	err = validator.CustomValidator().Struct("ENV-12400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageEnvironmentService.Update(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return nil, err
	}
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

	err = validator.CustomValidator().Struct("ENV-14400", &request)
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

	err = validator.CustomValidator().Struct("ENV-15400", &request)
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
