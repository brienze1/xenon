package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

const (
	mcpRelativePathReason     = "path must be an absolute file path"
	mcpBinaryBodySecretAction = "Saving a binary response body that contains a secret"
)

func (m *mcpExport) responseBody(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPSaveResponseBody
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !arguments.HasAbsolutePath() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpRelativePathReason)
	}

	target := arguments.ActivityTarget()
	last, err := loadMCPLastResult(ctx, m.manageResponseService, call.TabId())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	if entity.IsAIBinaryBody(last.IsBinary, last.ContentType) {
		err = m.saveBinaryBody(ctx, call, arguments)
	} else {
		err = m.saveTextBody(ctx, call, arguments, last)
	}
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}
	return dto.NewMCPToolResult(dto.NewMCPSavedPath(arguments.CleanPath())).WithTarget(target), nil
}

func (m *mcpExport) saveTextBody(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPSaveResponseBody, last dto.SendResult) error {
	text, err := readMCPResultText(ctx, m.errorHandler, m.manageResponseService, call.TabId(), last)
	if err != nil {
		return err
	}
	masked := guardMCPText(ctx, m.errorHandler, m.guardAIOutputService, call.LocalId, text)

	request := arguments.WriteAIOutputFile()
	err = validateWriteAIOutputFile(&request)
	if err != nil {
		return err
	}
	return m.writeAIOutputFileService.Write(ctx, request.FilePath(), []byte(masked))
}

func (m *mcpExport) saveBinaryBody(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPSaveResponseBody) error {
	raw, err := readMCPResponseText(ctx, m.errorHandler, m.manageResponseService, call.TabId())
	if err != nil {
		return err
	}
	if entity.AIOutputHasSecrets(raw, guardMCPText(ctx, m.errorHandler, m.guardAIOutputService, call.LocalId, raw)) {
		return mcpErrors.RejectHumanOnlyAIAction(mcpBinaryBodySecretAction)
	}

	request := arguments.SaveResponseBody(call.TabId())
	err = validateSaveResponseBody(&request)
	if err != nil {
		return err
	}
	return m.manageResponseService.SaveBody(ctx, request.TabId, request.GetPath())
}
