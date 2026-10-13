package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type console struct {
	errorHandler         adapter.ErrorHandler
	manageConsoleService appAdapter.ManageConsoleService
}

func NewConsole(errorHandler adapter.ErrorHandler, manageConsoleService appAdapter.ManageConsoleService) adapter.ConsoleController {
	return &console{errorHandler: errorHandler, manageConsoleService: manageConsoleService}
}

func (c *console) List(bindingCtx context.Context, request dto.ListConsole) (response []dto.ConsoleEntry, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CON-00400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewConsoleEntries(c.manageConsoleService.List(ctx, request.ToEntity()))

	logger.Info(ctx, "Finished", map[string]any{"count": len(result)})
	return result, nil
}

func (c *console) Get(bindingCtx context.Context, request dto.ConsoleEntryReference) (response dto.ConsoleEntry, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = dto.ConsoleEntry{}, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CON-02400", &request)
	if err != nil {
		return dto.ConsoleEntry{}, err
	}

	entry, err := c.manageConsoleService.Get(ctx, request.Id)
	if err != nil {
		return dto.ConsoleEntry{}, err
	}
	result := dto.NewConsoleEntryDetail(*entry)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (c *console) Clear(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = c.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	c.manageConsoleService.Clear(ctx)

	logger.Info(ctx, "Finished")
	return nil
}
