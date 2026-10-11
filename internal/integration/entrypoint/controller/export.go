package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type export struct {
	errorHandler  adapter.ErrorHandler
	exportService appAdapter.ExportService
}

func NewExport(errorHandler adapter.ErrorHandler, exportService appAdapter.ExportService) adapter.ExportController {
	return &export{errorHandler: errorHandler, exportService: exportService}
}

func (e *export) Options(bindingCtx context.Context, request dto.ExportItemReference) (response *dto.ExportOptions, err error) {
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

	err = validator.CustomValidator().Struct("EXP-10400", &request)
	if err != nil {
		return nil, err
	}

	options, err := e.exportService.Options(ctx, request.LocalId, request.Kind(), request.ItemId)
	if err != nil {
		return nil, err
	}
	result := dto.NewExportOptions(*options)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *export) PickTarget(bindingCtx context.Context, request dto.PickExportTarget) (response *dto.PickedPath, err error) {
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

	err = validator.CustomValidator().Struct("EXP-11400", &request)
	if err != nil {
		return nil, err
	}

	path, err := e.exportService.PickTarget(ctx, request.SuggestedFileName, request.ExportFormat(), request.ExportShape())
	if err != nil {
		return nil, err
	}
	result := dto.NewPickedPath(path)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *export) Export(bindingCtx context.Context, request dto.Export) (response *dto.ExportSummary, err error) {
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

	err = validator.CustomValidator().Struct("EXP-12400", &request)
	if err != nil {
		return nil, err
	}

	summary, err := e.exportService.Export(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewExportSummary(*summary)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *export) ShowInFolder(bindingCtx context.Context, request dto.ShowInFolder) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = e.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("EXP-13400", &request)
	if err != nil {
		return err
	}

	err = e.exportService.ShowInFolder(ctx, request.Path)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
