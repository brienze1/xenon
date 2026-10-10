package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type exampleManage struct {
	errorHandler         adapter.ErrorHandler
	manageExampleService appAdapter.ManageExampleService
}

func NewExampleManage(errorHandler adapter.ErrorHandler, manageExampleService appAdapter.ManageExampleService) adapter.ExampleManageController {
	return &exampleManage{errorHandler: errorHandler, manageExampleService: manageExampleService}
}

func (e *exampleManage) Get(bindingCtx context.Context, request dto.ExampleReference) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-05400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.Get(ctx, request.LocalId, request.RequestId, request.ExampleId)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) Create(bindingCtx context.Context, request dto.CreateExample) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-06400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.Create(ctx, request.LocalId, request.RequestId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) CreateFromResponse(bindingCtx context.Context, request dto.CreateExampleFromResponse) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-07400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.CreateFromResponse(ctx, request.LocalId, request.RequestId, request.TabId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) Save(bindingCtx context.Context, request dto.SaveExample) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-08400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.Save(ctx, request.LocalId, request.RequestId, request.ToEntity(), request.KeepsBody(), request.BaseRevision, request.Force)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) Rename(bindingCtx context.Context, request dto.RenameExample) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-09400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.Rename(ctx, request.LocalId, request.RequestId, request.ExampleId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) Duplicate(bindingCtx context.Context, request dto.ExampleReference) (response *dto.Example, err error) {
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

	err = validator.CustomValidator().Struct("EXM-10400", &request)
	if err != nil {
		return nil, err
	}

	details, err := e.manageExampleService.Duplicate(ctx, request.LocalId, request.RequestId, request.ExampleId)
	if err != nil {
		return nil, err
	}
	result := dto.NewExample(*details)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (e *exampleManage) Delete(bindingCtx context.Context, request dto.ExampleReference) (err error) {
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

	err = validator.CustomValidator().Struct("EXM-11400", &request)
	if err != nil {
		return err
	}

	if err = e.manageExampleService.Delete(ctx, request.LocalId, request.RequestId, request.ExampleId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
