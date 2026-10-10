package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type environmentSetActive struct {
	errorHandler                adapter.ErrorHandler
	setActiveEnvironmentService appAdapter.SetActiveEnvironmentService
}

func NewEnvironmentSetActive(errorHandler adapter.ErrorHandler, setActiveEnvironmentService appAdapter.SetActiveEnvironmentService) adapter.EnvironmentSetActiveController {
	return &environmentSetActive{errorHandler: errorHandler, setActiveEnvironmentService: setActiveEnvironmentService}
}

func (e *environmentSetActive) Handle(bindingCtx context.Context, request dto.SetActiveEnvironment) (response *dto.ActiveEnvironment, err error) {
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

	err = validator.CustomValidator().Struct("ENV-04400", &request)
	if err != nil {
		return nil, err
	}

	activeId, err := e.setActiveEnvironmentService.SetActive(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewActiveEnvironment(activeId)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
