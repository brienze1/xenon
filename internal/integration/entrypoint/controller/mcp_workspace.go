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
	mcpWorkspaceListAction   = "list"
	mcpWorkspaceActiveAction = "active"
)

var workspaceAITool = entity.AITool{
	Name:        "workspace",
	Description: "List workspaces or show the active one.",
	Actions:     []string{mcpWorkspaceListAction, mcpWorkspaceActiveAction},
	Effect:      aitooleffect.ReadOnly,
	Idempotent:  true,
}

type mcpWorkspace struct {
	errorHandler             adapter.ErrorHandler
	listWorkspacesService    appAdapter.ListWorkspacesService
	manageEnvironmentService appAdapter.ManageEnvironmentService
}

func NewMCPWorkspace(
	errorHandler adapter.ErrorHandler,
	listWorkspacesService appAdapter.ListWorkspacesService,
	manageEnvironmentService appAdapter.ManageEnvironmentService,
) adapter.MCPToolController {
	return &mcpWorkspace{
		errorHandler:             errorHandler,
		listWorkspacesService:    listWorkspacesService,
		manageEnvironmentService: manageEnvironmentService,
	}
}

func (m *mcpWorkspace) Tool() dto.MCPTool {
	return dto.NewMCPTool(workspaceAITool)
}

func (m *mcpWorkspace) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
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
	case mcpWorkspaceListAction:
		result, err = m.list(ctx)
	case mcpWorkspaceActiveAction:
		result, err = m.active(ctx)
	default:
		err = mcpErrors.RejectUnsupportedAIAction(call.Tool, call.Action)
	}
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpWorkspace) list(ctx context.Context) (dto.MCPToolResult, error) {
	registry, err := m.listWorkspacesService.List(ctx)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPWorkspaceList(*registry)), nil
}

func (m *mcpWorkspace) active(ctx context.Context) (dto.MCPToolResult, error) {
	registry, err := m.listWorkspacesService.List(ctx)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	workspace, found := registry.Active()
	if !found {
		return dto.MCPToolResult{}, mcpErrors.RejectInactiveAIWorkspace()
	}

	request := dto.NewListEnvironments(workspace.LocalID)
	err = validateListEnvironments(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	environments, err := m.manageEnvironmentService.List(ctx, request.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPActiveWorkspace(workspace, *environments)).WithTarget(workspace.NamedTarget()), nil
}
