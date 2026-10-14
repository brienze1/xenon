package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func guardMCPText(ctx context.Context, errorHandler adapter.ErrorHandler, guardAIOutputService appAdapter.GuardAIOutputService, localId, text string) (guarded string) {
	request := dto.NewAIGuard(localId, text)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			errorHandler.Handle(ctx, handlePanic(ctx, request, recovered), request)
			guarded = mcpErrors.UnguardedAIOutputText
		}
	}(ctx, request)

	if err := validateAIGuard(&request); err != nil {
		errorHandler.Handle(ctx, err, request)
		return mcpErrors.UnguardedAIOutputText
	}
	guarded, err := guardAIOutputService.Guard(ctx, request.WorkspaceId(), request.GuardedText())
	if err != nil {
		errorHandler.Handle(ctx, err, request)
	}
	return guarded
}
