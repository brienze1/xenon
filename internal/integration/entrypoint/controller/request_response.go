package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestResponse struct {
	errorHandler          adapter.ErrorHandler
	manageResponseService appAdapter.ManageResponseService
}

func NewRequestResponse(errorHandler adapter.ErrorHandler, manageResponseService appAdapter.ManageResponseService) adapter.RequestResponseController {
	return &requestResponse{errorHandler: errorHandler, manageResponseService: manageResponseService}
}

func (r *requestResponse) LastResult(bindingCtx context.Context, request dto.TabReference) (response *dto.SendResult, err error) {
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

	err = validator.CustomValidator().Struct("SND-01400", &request)
	if err != nil {
		return nil, err
	}

	last, err := r.manageResponseService.LastResult(ctx, request.TabId)
	if err != nil {
		return nil, err
	}
	result := dto.NewSendResult(request.TabId, *last)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestResponse) ReleaseResponse(bindingCtx context.Context, request dto.TabReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("SND-02400", &request)
	if err != nil {
		return err
	}

	if err = r.manageResponseService.Release(ctx, request.TabId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *requestResponse) SaveResponseBody(bindingCtx context.Context, request dto.SaveResponseBody) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("SND-04400", &request)
	if err != nil {
		return err
	}

	if err = r.manageResponseService.SaveBody(ctx, request.TabId, request.GetPath()); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *requestResponse) OpenBody(bindingCtx context.Context, request dto.TabReference) (response *dto.ResponseBody, err error) {
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

	err = validator.CustomValidator().Struct("SND-14400", &request)
	if err != nil {
		return nil, err
	}

	found, content, err := r.manageResponseService.OpenBody(ctx, request.TabId)
	if err != nil {
		return nil, err
	}
	result := dto.NewResponseBody(*found, content)

	logger.Info(ctx, "Finished", request)
	return &result, nil
}

func (r *requestResponse) ReleaseAll(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	if err = r.manageResponseService.ReleaseAll(ctx); err != nil {
		return err
	}

	logger.Info(ctx, "Finished")
	return nil
}
