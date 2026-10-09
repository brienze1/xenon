package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestEdit struct {
	errorHandler       adapter.ErrorHandler
	editRequestService appAdapter.EditRequestService
}

func NewRequestEdit(errorHandler adapter.ErrorHandler, editRequestService appAdapter.EditRequestService) adapter.RequestEditController {
	return &requestEdit{errorHandler: errorHandler, editRequestService: editRequestService}
}

func (r *requestEdit) ApplyUrl(bindingCtx context.Context, request dto.ApplyRequest) (response *dto.Request, err error) {
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

	err = validator.CustomValidator().Struct("REQ-03400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewRequest(r.editRequestService.ApplyUrl(ctx, request.Request.ToEntity()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestEdit) ApplyParams(bindingCtx context.Context, request dto.ApplyRequest) (response *dto.Request, err error) {
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

	err = validator.CustomValidator().Struct("REQ-04400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewRequest(r.editRequestService.ApplyParams(ctx, request.Request.ToEntity()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestEdit) ParseBulk(bindingCtx context.Context, request dto.BulkText) (response *dto.BulkRows, err error) {
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

	err = validator.CustomValidator().Struct("REQ-05400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewBulkRows(r.editRequestService.ParseBulk(ctx, request.Text))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestEdit) FormatBulk(bindingCtx context.Context, request dto.BulkRows) (response *dto.BulkText, err error) {
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

	err = validator.CustomValidator().Struct("REQ-06400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewBulkText(r.editRequestService.FormatBulk(ctx, request.ToEntity()))

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *requestEdit) AutoHeaders(bindingCtx context.Context, request dto.AutoHeadersRequest) (response dto.AutoHeaderList, err error) {
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

	err = validator.CustomValidator().Struct("REQ-07400", &request)
	if err != nil {
		return nil, err
	}

	headers, err := r.editRequestService.AutoHeaders(ctx, request.LocalId, request.Request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewAutoHeaderList(headers)

	logger.Info(ctx, "Finished", result)
	return result, nil
}
