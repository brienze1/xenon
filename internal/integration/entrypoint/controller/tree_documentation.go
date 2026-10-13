package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type treeDocumentation struct {
	errorHandler               adapter.ErrorHandler
	manageDocumentationService appAdapter.ManageDocumentationService
	autoSyncService            appAdapter.AutoSyncService
}

func NewTreeDocumentation(errorHandler adapter.ErrorHandler, manageDocumentationService appAdapter.ManageDocumentationService, autoSyncService appAdapter.AutoSyncService) adapter.TreeDocumentationController {
	return &treeDocumentation{errorHandler: errorHandler, manageDocumentationService: manageDocumentationService, autoSyncService: autoSyncService}
}

func (t *treeDocumentation) GetDocumentation(bindingCtx context.Context, request dto.TreeItemReference) (response *dto.Documentation, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DOC-11400", &request)
	if err != nil {
		return nil, err
	}

	documentation, err := t.manageDocumentationService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewDocumentation(*documentation)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeDocumentation) PreviewDocumentation(bindingCtx context.Context, request dto.DocumentationInput) (response *dto.Documentation, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DOC-12400", &request)
	if err != nil {
		return nil, err
	}

	documentation, err := t.manageDocumentationService.Preview(ctx, request.Yaml)
	if err != nil {
		return nil, err
	}
	result := dto.NewDocumentation(*documentation)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeDocumentation) PlanDocumentation(bindingCtx context.Context, request dto.DocumentationInput) (response *dto.DocumentationPlan, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DOC-13400", &request)
	if err != nil {
		return nil, err
	}

	plan, err := t.manageDocumentationService.Plan(ctx, request.LocalId, request.Id, request.Yaml)
	if err != nil {
		return nil, err
	}
	result := dto.NewDocumentationPlan(*plan)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeDocumentation) ApplyDocumentation(bindingCtx context.Context, request dto.DocumentationInput) (response *dto.DocumentationApplied, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("DOC-14400", &request)
	if err != nil {
		return nil, err
	}

	applied, err := t.manageDocumentationService.Apply(ctx, request.LocalId, request.Id, request.Yaml, request.ConfirmDeletions, request.KeepDeleted)
	if err != nil {
		return nil, err
	}
	triggerAutoSync(ctx, t.autoSyncService.Saved, request.LocalId)
	result := dto.NewDocumentationApplied(*applied)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
