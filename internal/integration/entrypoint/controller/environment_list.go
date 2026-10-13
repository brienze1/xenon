package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type environmentList struct {
	errorHandler             adapter.ErrorHandler
	manageEnvironmentService appAdapter.ManageEnvironmentService
}

func NewEnvironmentList(errorHandler adapter.ErrorHandler, manageEnvironmentService appAdapter.ManageEnvironmentService) adapter.EnvironmentListController {
	return &environmentList{errorHandler: errorHandler, manageEnvironmentService: manageEnvironmentService}
}

func (e *environmentList) Handle(bindingCtx context.Context, request dto.ListEnvironments) (response *dto.EnvironmentList, err error) {
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

	err = validator.CustomValidator().Struct("ENV-03400", &request)
	if err != nil {
		return nil, err
	}

	list, err := e.manageEnvironmentService.List(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewEnvironmentList(*list)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
