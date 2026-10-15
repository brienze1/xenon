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
	mcpHistoryListAction = "list"
	mcpHistoryGetAction  = "get"
)

var historyAITool = entity.AITool{
	Name:            "history",
	Description:     "List and read History (sends and AI activity).",
	Actions:         []string{mcpHistoryListAction, mcpHistoryGetAction},
	Effect:          aitooleffect.ReadOnly,
	Idempotent:      true,
	WorkspaceScoped: true,
}

type mcpHistory struct {
	errorHandler         adapter.ErrorHandler
	manageHistoryService appAdapter.ManageHistoryService
}

func NewMCPHistory(errorHandler adapter.ErrorHandler, manageHistoryService appAdapter.ManageHistoryService) adapter.MCPToolController {
	return &mcpHistory{errorHandler: errorHandler, manageHistoryService: manageHistoryService}
}

func (m *mcpHistory) Tool() dto.MCPTool {
	return dto.NewMCPTool(
		historyAITool,
		dto.NewMCPToolArguments[dto.MCPHistoryList](mcpHistoryListAction),
		dto.NewMCPToolArguments[dto.MCPHistoryGet](mcpHistoryGetAction),
	)
}

func (m *mcpHistory) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpHistoryListAction:
		result, err = m.list(ctx, call)
	case mcpHistoryGetAction:
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

func (m *mcpHistory) list(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPHistoryList
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.ListHistory(call.LocalId)
	err = validateListHistory(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entries, err := m.manageHistoryService.List(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPPage(entity.NewAIHistoryPage(entries), dto.NewMCPHistoryLine)), nil
}

func (m *mcpHistory) get(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPHistoryGet
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request := arguments.HistoryReference(call.LocalId)
	err = validateGetHistory(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	entry, err := m.manageHistoryService.Get(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if entry.IsAction() {
		return dto.NewMCPToolResult(dto.NewMCPHistoryAction(*entry.Activity)).WithTarget(entry.ActivityTarget()), nil
	}
	return dto.NewMCPToolResult(dto.NewMCPHistorySend(*entry)).WithTarget(entry.ActivityTarget()), nil
}
