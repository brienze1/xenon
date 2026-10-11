package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type snippet struct {
	errorHandler           adapter.ErrorHandler
	generateSnippetService appAdapter.GenerateSnippetService
}

func NewSnippet(errorHandler adapter.ErrorHandler, generateSnippetService appAdapter.GenerateSnippetService) adapter.SnippetController {
	return &snippet{errorHandler: errorHandler, generateSnippetService: generateSnippetService}
}

func (s *snippet) Languages(bindingCtx context.Context) (response dto.SnippetLanguageList, err error) {
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

	result := dto.NewSnippetLanguageList(s.generateSnippetService.Languages(ctx))

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (s *snippet) Generate(bindingCtx context.Context, request dto.GenerateSnippet) (response *dto.Snippet, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, s.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("SNP-00400", &request)
	if err != nil {
		return nil, err
	}

	generated, err := s.generateSnippetService.Generate(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewSnippet(*generated)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (s *snippet) Copy(bindingCtx context.Context, request dto.GenerateSnippet) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = s.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("SNP-01400", &request)
	if err != nil {
		return err
	}

	err = s.generateSnippetService.Copy(ctx, request.ToEntity())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
