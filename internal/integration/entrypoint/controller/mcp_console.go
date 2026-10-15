package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	mcpConsoleListAction = "list"
	mcpConsoleGetAction  = "get"
	mcpConsoleTypeReason = "type must be request, log or error"
)

var consoleAITool = entity.AITool{
	Name:            "console",
	Description:     "Read console entries from sends and scripts.",
	Actions:         []string{mcpConsoleListAction, mcpConsoleGetAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpConsole struct {
	errorHandler         adapter.ErrorHandler
	manageConsoleService appAdapter.ManageConsoleService
}

func NewMCPConsole(errorHandler adapter.ErrorHandler, manageConsoleService appAdapter.ManageConsoleService) adapter.MCPToolController {
	return &mcpConsole{errorHandler: errorHandler, manageConsoleService: manageConsoleService}
}

func (m *mcpConsole) Tool() dto.MCPTool {
	return dto.NewMCPTool(
		consoleAITool,
		dto.NewMCPToolArguments[dto.MCPConsoleList](mcpConsoleListAction),
		dto.NewMCPToolArguments[dto.MCPConsoleGet](mcpConsoleGetAction),
	)
}

func (m *mcpConsole) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpConsoleListAction:
		result, err = m.list(ctx, call)
	case mcpConsoleGetAction:
		result, err = m.get(ctx, call)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpConsole) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPConsoleList
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	request, known := arguments.ListConsole()
	if !known {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpConsoleTypeReason)
	}

	err = validateListConsole(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entries := m.manageConsoleService.List(ctx, request.ToEntity())
	page, valid := arguments.Page(entries)
	if !valid {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpUnknownCursorReason)
	}
	return dto.NewMCPToolResult(dto.NewMCPPage(page, dto.NewMCPConsoleLine)), nil
}

func (m *mcpConsole) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPConsoleGet
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.ConsoleEntryReference()
	err = validateGetConsoleEntry(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entry, err := m.manageConsoleService.Get(ctx, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewConsoleEntryDetail(*entry)).WithTarget(entry.ActivityTarget()), nil
}
