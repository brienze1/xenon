package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type scriptTools struct {
	errorHandler       adapter.ErrorHandler
	scriptToolsService appAdapter.ScriptToolsService
}

func NewScriptTools(errorHandler adapter.ErrorHandler, scriptToolsService appAdapter.ScriptToolsService) adapter.ScriptToolsController {
	return &scriptTools{errorHandler: errorHandler, scriptToolsService: scriptToolsService}
}

func (s *scriptTools) Check(bindingCtx context.Context, request dto.CheckScript) (response dto.ScriptCheck, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = dto.ScriptCheck{}, s.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	result := dto.NewScriptCheck(s.scriptToolsService.Check(ctx, request.Text))

	logger.Info(ctx, "Finished", map[string]any{"errors": len(result.Errors)})
	return result, nil
}

func (s *scriptTools) Reference(bindingCtx context.Context) (response dto.ScriptReference, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = dto.ScriptReference{}, s.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	members, err := s.scriptToolsService.Reference(ctx)
	if err != nil {
		return dto.ScriptReference{}, err
	}
	result := dto.NewScriptReference(members)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (s *scriptTools) Snippets(bindingCtx context.Context) (response []dto.ScriptSnippet, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, s.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	snippets, err := s.scriptToolsService.Snippets(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewScriptSnippets(snippets)

	logger.Info(ctx, "Finished", map[string]any{"snippets": len(result)})
	return result, nil
}
