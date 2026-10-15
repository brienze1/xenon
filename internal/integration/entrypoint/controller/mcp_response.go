package controller

import (
	"context"
	"strconv"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpResponseLastAction = "last"
	mcpResponseBodyAction = "body"
)

var mcpBodyRangeReason = "offset must be 0 or more and length at most " + strconv.Itoa(entity.AIMaxBodyReadBytes)

var responseAITool = entity.AITool{
	Name:            "response",
	Description:     "Read the last response of an AI send, including more of a truncated body.",
	Actions:         []string{mcpResponseLastAction, mcpResponseBodyAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpResponse struct {
	errorHandler          adapter.ErrorHandler
	manageResponseService appAdapter.ManageResponseService
	guardAIOutputService  appAdapter.GuardAIOutputService
}

func NewMCPResponse(
	errorHandler adapter.ErrorHandler,
	manageResponseService appAdapter.ManageResponseService,
	guardAIOutputService appAdapter.GuardAIOutputService,
) adapter.MCPToolController {
	return &mcpResponse{
		errorHandler:          errorHandler,
		manageResponseService: manageResponseService,
		guardAIOutputService:  guardAIOutputService,
	}
}

func (m *mcpResponse) Tool() dto.MCPTool {
	return dto.NewMCPTool(responseAITool, dto.NewMCPToolArguments[dto.MCPResponseBody](mcpResponseBodyAction))
}

func (m *mcpResponse) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
	ctx := logger.GetContext(toolCtx)
	logger.Info(ctx, "Started", call)

	defer func(ctx context.Context, call any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, call, recovered)
		}
		if err != nil {
			err = m.errorHandler.Handle(ctx, err, call)
		}
	}(ctx, call)

	switch call.Action {
	case mcpResponseLastAction:
		result, err = m.last(ctx, call)
	case mcpResponseBodyAction:
		result, err = m.body(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpResponse) last(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	last, err := loadMCPLastResult(ctx, m.manageResponseService, call.TabId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	response, err := newMCPSendResponse(ctx, m.errorHandler, m.manageResponseService, m.guardAIOutputService, call, last)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(response), nil
}

func (m *mcpResponse) body(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPResponseBody
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	last, err := loadMCPLastResult(ctx, m.manageResponseService, call.TabId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if entity.IsAIBinaryBody(last.IsBinary, last.ContentType) {
		return dto.NewMCPToolResult(dto.NewMCPBinaryBody(last)), nil
	}

	text, err := readMCPResultText(ctx, m.errorHandler, m.manageResponseService, call.TabId(), last)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	masked := guardMCPText(ctx, m.errorHandler, m.guardAIOutputService, call.LocalId, text)
	slice, valid := arguments.Slice(masked)
	if !valid {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpBodyRangeReason)
	}
	return dto.NewMCPToolResult(dto.NewMCPBodySlice(slice)), nil
}
