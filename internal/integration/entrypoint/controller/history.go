package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type history struct {
	errorHandler   adapter.ErrorHandler
	historyService appAdapter.HistoryService
}

func NewHistory(errorHandler adapter.ErrorHandler, historyService appAdapter.HistoryService) adapter.HistoryController {
	return &history{errorHandler: errorHandler, historyService: historyService}
}

func (h *history) List(bindingCtx context.Context, request dto.ListHistory) (response dto.HistoryEntryList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, h.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("HIS-00400", &request)
	if err != nil {
		return nil, err
	}

	entries, err := h.historyService.List(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewHistoryEntryList(entries)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (h *history) Get(bindingCtx context.Context, request dto.HistoryReference) (response *dto.HistoryEntryDetail, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, h.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("HIS-01400", &request)
	if err != nil {
		return nil, err
	}

	entry, err := h.historyService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewHistoryEntryDetail(*entry)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (h *history) Clear(bindingCtx context.Context, request dto.ClearHistory) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = h.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("HIS-02400", &request)
	if err != nil {
		return err
	}

	err = h.historyService.Clear(ctx, request.LocalId)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
