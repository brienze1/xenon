package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestGet struct {
	errorHandler       adapter.ErrorHandler
	saveRequestService appAdapter.SaveRequestService
}

func NewRequestGet(errorHandler adapter.ErrorHandler, saveRequestService appAdapter.SaveRequestService) adapter.RequestGetController {
	return &requestGet{errorHandler: errorHandler, saveRequestService: saveRequestService}
}

func (r *requestGet) Get(bindingCtx context.Context, request dto.OpenRequest) (response *dto.Request, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("REQ-08400", &request)
	if err != nil {
		return nil, err
	}

	saved, err := r.saveRequestService.Open(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
