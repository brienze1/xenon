package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestSend struct {
	errorHandler       adapter.ErrorHandler
	sendRequestService appAdapter.SendRequestService
}

func NewRequestSend(errorHandler adapter.ErrorHandler, sendRequestService appAdapter.SendRequestService) adapter.RequestSendController {
	return &requestSend{errorHandler: errorHandler, sendRequestService: sendRequestService}
}

func (r *requestSend) Send(bindingCtx context.Context, request dto.SendRequest) (response *dto.SendResult, err error) {
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

	err = validateSendRequest(&request)
	if err != nil {
		return nil, err
	}

	sent, err := r.sendRequestService.Send(ctx, request.WorkspaceId(), request.TargetTabId(), request.ToEntity(), properties.Properties().Application.Version, request.ToAcknowledgement(), request.ToSendOrigin())
	if err != nil {
		return nil, err
	}
	result := dto.NewSendResult(request.TargetTabId(), *sent)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
