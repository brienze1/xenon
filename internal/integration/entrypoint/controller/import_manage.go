package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type importManage struct {
	errorHandler        adapter.ErrorHandler
	manageImportService appAdapter.ManageImportService
}

func NewImportManage(errorHandler adapter.ErrorHandler, manageImportService appAdapter.ManageImportService) adapter.ImportManageController {
	return &importManage{errorHandler: errorHandler, manageImportService: manageImportService}
}

func (i *importManage) Detect(bindingCtx context.Context, request dto.DetectImport) (response *dto.ImportPreview, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, i.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("IMP-09400", &request)
	if err != nil {
		return nil, err
	}

	source, err := i.manageImportService.Detect(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewImportPreview(*source)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (i *importManage) PickSources(bindingCtx context.Context) (response *dto.PickedSources, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, i.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	paths, err := i.manageImportService.PickSources(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewPickedSources(paths)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (i *importManage) Clashes(bindingCtx context.Context, request dto.ImportPlan) (response *dto.ImportClashes, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, i.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("IMP-12400", &request)
	if err != nil {
		return nil, err
	}

	clashes, err := i.manageImportService.Clashes(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewImportClashes(clashes)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (i *importManage) Import(bindingCtx context.Context, request dto.ImportPlan) (response *dto.ImportResult, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, i.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("IMP-13400", &request)
	if err != nil {
		return nil, err
	}

	imported, err := i.manageImportService.Import(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewImportResult(*imported)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (i *importManage) Discard(bindingCtx context.Context, request dto.ImportSourceReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = i.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("IMP-14400", &request)
	if err != nil {
		return err
	}

	if err = i.manageImportService.Discard(ctx, request.SourceId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
