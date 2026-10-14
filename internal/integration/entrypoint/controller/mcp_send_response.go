package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func newMCPSendResponse(
	ctx context.Context,
	errorHandler adapter.ErrorHandler,
	manageResponseService appAdapter.ManageResponseService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	call dto.MCPToolCall,
	result dto.SendResult,
) (dto.MCPSendResponse, error) {
	response := dto.NewMCPSendResponse(result)
	if !result.HasAITextBody() {
		return response, nil
	}

	text, err := readMCPResultText(ctx, errorHandler, manageResponseService, call.TabId(), result)
	if err != nil {
		return response, err
	}
	return response.WithBody(entity.NewAIInlineBody(guardMCPText(ctx, errorHandler, guardAIOutputService, call.LocalId, text))), nil
}
