package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type variableInspect struct {
	errorHandler            adapter.ErrorHandler
	inspectVariablesService appAdapter.InspectVariablesService
	suggestVariablesService appAdapter.SuggestVariablesService
}

func NewVariableInspect(errorHandler adapter.ErrorHandler, inspectVariablesService appAdapter.InspectVariablesService, suggestVariablesService appAdapter.SuggestVariablesService) adapter.VariableInspectController {
	return &variableInspect{errorHandler: errorHandler, inspectVariablesService: inspectVariablesService, suggestVariablesService: suggestVariablesService}
}

func (v *variableInspect) Inspect(bindingCtx context.Context, request dto.InspectVariables) (response *dto.VariableInspection, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-15400", &request)
	if err != nil {
		return nil, err
	}

	inspection, err := v.inspectVariablesService.Inspect(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableInspection(*inspection)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableInspect) Suggest(bindingCtx context.Context, request dto.SuggestVariables) (response dto.VariableSuggestionList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-16400", &request)
	if err != nil {
		return nil, err
	}

	suggestions, err := v.suggestVariablesService.Suggest(ctx, request.LocalId, request.RequestId, request.Prefix)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableSuggestionList(suggestions)

	logger.Info(ctx, "Finished", result)
	return result, nil
}
