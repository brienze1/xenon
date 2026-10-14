package controller

import (
	"context"
	"io"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func loadMCPLastResult(ctx context.Context, manageResponseService appAdapter.ManageResponseService, tabId string) (dto.SendResult, error) {
	request := dto.NewTabReference(tabId)
	err := validateLastResponse(&request)
	if err != nil {
		return dto.SendResult{}, err
	}

	last, err := manageResponseService.LastResult(ctx, request.TabId)
	if err != nil {
		return dto.SendResult{}, err
	}
	return dto.NewLastSendResult(request.TabId, *last), nil
}

func readMCPResultText(ctx context.Context, errorHandler adapter.ErrorHandler, manageResponseService appAdapter.ManageResponseService, tabId string, result dto.SendResult) (string, error) {
	if text, inline := result.InlineText(); inline {
		return text, nil
	}
	return readMCPResponseText(ctx, errorHandler, manageResponseService, tabId)
}

func readMCPResponseText(ctx context.Context, errorHandler adapter.ErrorHandler, manageResponseService appAdapter.ManageResponseService, tabId string) (string, error) {
	request := dto.NewTabReference(tabId)
	err := validateOpenResponseBody(&request)
	if err != nil {
		return "", err
	}

	found, content, err := manageResponseService.OpenBody(ctx, request.TabId)
	if err != nil {
		return "", err
	}
	body := dto.NewResponseBody(*found, content)
	defer closeMCPResponseBody(ctx, errorHandler, body)

	text, err := io.ReadAll(body.Content)
	if err != nil {
		return "", mcpErrors.FailUnknownAICall(err)
	}
	return string(text), nil
}

func closeMCPResponseBody(ctx context.Context, errorHandler adapter.ErrorHandler, body dto.ResponseBody) {
	if err := body.Content.Close(); err != nil {
		errorHandler.Handle(ctx, err, body)
	}
}
