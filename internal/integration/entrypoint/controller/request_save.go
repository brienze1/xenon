package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestSave struct {
	errorHandler       adapter.ErrorHandler
	saveRequestService appAdapter.SaveRequestService
}

func NewRequestSave(errorHandler adapter.ErrorHandler, saveRequestService appAdapter.SaveRequestService) adapter.RequestSaveController {
	return &requestSave{errorHandler: errorHandler, saveRequestService: saveRequestService}
}

func (r *requestSave) Save(bindingCtx context.Context, request dto.SaveRequest) (response *dto.Request, err error) {
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

	err = validator.CustomValidator().Struct("REQ-09400", &request)
	if err != nil {
		return nil, err
	}

	saved, err := r.saveRequestService.Save(ctx, request.LocalId, request.Request.ToEntity(), request.BaseRevision, request.Force)
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestSave) SaveAs(bindingCtx context.Context, request dto.SaveRequestAs) (response *dto.Request, err error) {
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

	err = validator.CustomValidator().Struct("REQ-10400", &request)
	if err != nil {
		return nil, err
	}

	saved, err := r.saveRequestService.SaveAs(ctx, request.LocalId, request.ParentId, request.Name, request.Request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewRequest(*saved)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
